package wallet

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/atlasdev/orbitron/internal/config"
)

const (
	// Tokens
	USDCeAddress            = "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174"
	ConditionalTokenAddress = "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045"

	// Spenders
	MainExchangeAddress    = "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E"
	NegRiskExchangeAddress = "0xC5d563A36AE78145C45a50134d48A1215220f80a"
	NegRiskAdapterAddress  = "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296"
)

// AllowanceChecker is a subset of ethclient for testing.
type AllowanceChecker interface {
	ethereum.ContractCaller
	ethereum.ChainReader
	ethereum.TransactionSender
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	ChainID(ctx context.Context) (*big.Int, error)
}

// CheckAllowances checks the allowances for USDC and Conditional Tokens for the three Polymarket contracts.
func CheckAllowances(ctx context.Context, rpcURL, ownerAddr string) ([]config.AllowanceStatus, error) {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial rpc: %w", err)
	}
	defer client.Close()

	return CheckAllowancesWithClient(ctx, client, ownerAddr)
}

// CheckAllowancesWithClient is the testable implementation of CheckAllowances.
func CheckAllowancesWithClient(ctx context.Context, client AllowanceChecker, ownerAddr string) ([]config.AllowanceStatus, error) {
	owner := common.HexToAddress(ownerAddr)

	tokens := []struct {
		Symbol  string
		Address string
	}{
		{"USDC.e", USDCeAddress},
		{"CTF", ConditionalTokenAddress},
	}

	spenders := []struct {
		Name    string
		Address string
	}{
		{"Main Exchange", MainExchangeAddress},
		{"Neg Risk Exchange", NegRiskExchangeAddress},
		{"Neg Risk Adapter", NegRiskAdapterAddress},
	}

	var results []config.AllowanceStatus
	for _, token := range tokens {
		tokenAddr := common.HexToAddress(token.Address)
		for _, spender := range spenders {
			spenderAddr := common.HexToAddress(spender.Address)
			approved, err := isApproved(ctx, client, tokenAddr, owner, spenderAddr)
			if err != nil {
				approved = false
			}

			results = append(results, config.AllowanceStatus{
				TokenSymbol: token.Symbol,
				Token:       token.Address,
				SpenderName: spender.Name,
				Spender:     spender.Address,
				Approved:    approved,
			})
		}
	}

	return results, nil
}

// GrantMissingAllowances automatically approves tokens for spenders if allowance is missing.
func GrantMissingAllowances(ctx context.Context, rpcURL, privKey string, statuses []config.AllowanceStatus) error {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return fmt.Errorf("dial rpc: %w", err)
	}
	defer client.Close()

	return GrantMissingAllowancesWithClient(ctx, client, privKey, statuses)
}

// GrantMissingAllowancesWithClient is the testable implementation of GrantMissingAllowances.
func GrantMissingAllowancesWithClient(ctx context.Context, client AllowanceChecker, privKey string, statuses []config.AllowanceStatus) error {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(privKey, "0x"))
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("get chain id: %w", err)
	}

	fromAddr := crypto.PubkeyToAddress(key.PublicKey)
	nonce, err := client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return fmt.Errorf("get nonce: %w", err)
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		gasPrice = big.NewInt(50000000000) // fallback 50 gwei
	}

	for _, s := range statuses {
		if s.Approved {
			continue
		}

		spender := common.HexToAddress(s.Spender)
		token := common.HexToAddress(s.Token)

		maxUint256, _ := new(big.Int).SetString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16)

		data := make([]byte, 0, 4+32*2)
		methodID := crypto.Keccak256([]byte("approve(address,uint256)"))[:4]
		data = append(data, methodID...)
		data = append(data, common.LeftPadBytes(spender.Bytes(), 32)...)
		data = append(data, common.LeftPadBytes(maxUint256.Bytes(), 32)...)

		tx := types.NewTransaction(nonce, token, big.NewInt(0), 100000, gasPrice, data)
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), key)
		if err != nil {
			return fmt.Errorf("sign tx: %w", err)
		}

		if err := client.SendTransaction(ctx, signedTx); err != nil {
			return fmt.Errorf("send tx for %s: %w (ensure wallet has MATIC for gas)", s.TokenSymbol, err)
		}

		nonce++
	}

	return nil
}

// isApproved checks if the allowance is sufficient (non-zero).
// Polymarket usually wants "infinite" allowance, but any non-zero value often suffices to start.
func isApproved(ctx context.Context, client AllowanceChecker, token, owner, spender common.Address) (bool, error) {
	// allowance(address owner, address spender)
	data := make([]byte, 0, 4+32*2)
	methodID := crypto.Keccak256([]byte("allowance(address,address)"))[:4]
	data = append(data, methodID...)
	data = append(data, common.LeftPadBytes(owner.Bytes(), 32)...)
	data = append(data, common.LeftPadBytes(spender.Bytes(), 32)...)

	msg := ethereum.CallMsg{
		To:   &token,
		Data: data,
	}

	res, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		return false, err
	}

	if len(res) == 0 {
		return false, fmt.Errorf("empty response")
	}

	allowance := new(big.Int).SetBytes(res)
	return allowance.Cmp(big.NewInt(0)) > 0, nil
}
