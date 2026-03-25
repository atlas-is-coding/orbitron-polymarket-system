package auth

import (
	"fmt"
)

// ContractConfig contains addresses for Polymarket contracts.
type ContractConfig struct {
	Exchange          string
	Collateral        string
	ConditionalTokens string
}

var (
	// Standard contract configurations by ChainID.
	Config = map[int64]ContractConfig{
		137: {
			Exchange:          "0x4bFb41d5B3570DeFd03C39a9A4D8dE6Bd8B8982E",
			Collateral:        "0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174",
			ConditionalTokens: "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045",
		},
		80002: {
			Exchange:          "0xdFE02Eb6733538f8Ea35D585af8DE5958AD99E40",
			Collateral:        "0x9c4e1703476e875070ee25b56a58b008cfb8fa78",
			ConditionalTokens: "0x69308FB512518e39F9b16112fA8d994F4e2Bf8bB",
		},
	}

	// Negative Risk contract configurations by ChainID.
	NegRiskConfig = map[int64]ContractConfig{
		137: {
			Exchange:          "0xC5d563A36AE78145C45a50134d48A1215220f80a",
			Collateral:        "0x2791bca1f2de4661ed88a30c99a7a944984174",
			ConditionalTokens: "0x4D97DCd97eC945f40cF65F87097ACe5EA0476045",
		},
		80002: {
			Exchange:          "0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296",
			Collateral:        "0x9c4e1703476e875070ee25b56a58b008cfb8fa78",
			ConditionalTokens: "0x69308FB512518e39F9b16112fA8d994F4e2Bf8bB",
		},
	}
)

// GetContractConfig returns the contract configuration for the specified chain and risk type.
func GetContractConfig(chainID int64, negRisk bool) (ContractConfig, error) {
	var cfg ContractConfig
	var ok bool

	if negRisk {
		cfg, ok = NegRiskConfig[chainID]
	} else {
		cfg, ok = Config[chainID]
	}

	if !ok {
		return ContractConfig{}, fmt.Errorf("invalid or unsupported chainID: %d", chainID)
	}

	return cfg, nil
}
