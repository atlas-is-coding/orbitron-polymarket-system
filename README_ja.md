# Polytrade Bot 📈🤖

*他の言語で読む: [English](README.md), [Русский](README_ru.md), [中文](README_zh.md), [한국어](README_ko.md).*

Polytrade Botは、**Polymarket CTF Exchange**向けの高度なアルゴリズム取引および管理ボットです。インタラクティブなターミナルユーザーインターフェース（TUI）、Vue 3によるWeb UI、遠隔管理用のTelegramボットなど、堅牢なマルチインターフェースアーキテクチャを備えています。

## 🌟 詳細な機能

*   **マルチインターフェース体験:**
    *   **TUI:** 市場、取引、コピートレード、ウォレット、戦略、設定、ログ用の個別のタブを備えた美しいターミナルインターフェース。
    *   **Web UI:** リアルタイムWebSocket更新、JWT認証、レスポンシブデザインを備えたVue 3 SPA。
    *   **Telegram Bot:** TUIの機能を完全にミラーリングする、インラインキーボードと複数ステップの会話機能を備えたインタラクティブなボット。
*   **アルゴリズム取引エンジン:** アービトラージ（裁定取引）、クロスマーケット、Fade Chaos、マーケットメイキングなどの戦略を内蔵。カスタム戦略を簡単に登録して追加できます。
*   **高度なコピートレード:** Data APIを介してターゲットのウォレットをリアルタイムで監視し、CLOB API経由でポジションを自動的にコピーします。動的なボリューム割り当てモード（`proportional` または `fixed_pct`）をサポートしています。
*   **リアルタイム監視とアラート:**
    *   **取引モニター:** 未約定の注文、取引の実行、ポジションを追跡します。
    *   **市場アラート:** 市場状態の変化に基づいてアラート条件をリアルタイムで評価します。
*   **安全な認証システム:** L1/L2クレデンシャルアーキテクチャ。EIP-712署名を自動的に生成し、L2キーはメモリ内でのみ保持され、設定ファイルに保存されることは絶対にありません。セキュリティを最大限に高めるため、署名は30秒後に自動的に期限切れになります。
*   **マルチウォレット対応:** UIから直接複数のアクティブなウォレットを管理し、オン/オフを切り替え、集計された統計を表示できます。
*   **多言語対応（i18n）:** 英語、ロシア語、中国語、日本語、韓国語をネイティブにサポートし、再起動なしで即座に切り替え可能です。

## 🏗 アーキテクチャの概要

ボットは、コンテキストでキャンセル可能な7つのコアサブシステムで動作します：

1.  **WebSocket クライアント:** Polymarket CLOB（`market`、`user`、`asset`チャネル）への自動再接続機能を備えた永続的な接続。
2.  **モニター:** Gamma APIとData APIをポーリングし、市場の変化を評価してアラートをトリガーします。
3.  **取引エンジン:** プラグイン可能な取引戦略（`trading.Strategy`）を実行するための、拡張性の高いGoroutineベースのレイヤー。
4.  **通知システム:** 設定可能なアラートシステム（デフォルトはTelegram）。
5.  **コピートレーダー:** 設定されたウォレットを追跡し、ポジションを複製します。ボットを再起動することなく、設定変更のホットリロードをサポートします。
6.  **Telegram Bot:** 単一管理者モデル（`AdminChatID`）を使用したTUIの対話型ミラー。
7.  **Web UI:** Vue 3 SPAを提供する組み込みHTTPサーバー + WebSocketハブ。

## ⚙️ 構成 (`config.toml`)

ボットはすべて`config.toml`によって制御されます。安全のため、取引とデータベース機能はデフォルトで無効になっています。

主要なセクション：
*   `[auth]`: `private_key`（16進数、`0x`プレフィックスなし）が必要です。L2クレデンシャルは起動時に自動的に派生します。
*   `[webui]`: `enabled` (true/false)、`listen` (例：`127.0.0.1:8080`)、`jwt_secret` (トークンの署名とログインパスワードとして使用)。
*   `[ui]`: `language` (`en`、`ru`、`zh`、`ja`、`ko`)。即座にホットリロードされます。
*   `[monitor.trades]`: `enabled`、`poll_interval_ms`。L2認証が必要です。
*   `[copytrading]`: `enabled`、`size_mode` (`proportional`/`fixed_pct`)、および`[[copytrading.traders]]`リスト。データベースとL2が必要です。
*   `[telegram]`: `enabled`、`bot_token`、`admin_chat_id` (単一の管理者対象)。
*   `[database]`: `enabled`、`path` (SQLite DBのパス)。
*   `chain_id`: Polygonメインネットの場合は`137`、Amoyテストネットの場合は`80002`。

## 🚀 インストールとセットアップ

### 必須環境

*   [Go 1.24+](https://golang.org/doc/install)
*   [Node.js 18+](https://nodejs.org/) (Web UIを変更する場合のみ必要)
*   Polymarketウォレットのプライベートキー

### セットアップ手順

#### オプション 1: ユニバーサルセットアップスクリプト (推奨)
Linux、macOS、およびWindows (Git Bash/WSL経由) で動作するユニバーサルな `setup.sh` スクリプトを提供しています。これにより、GoとNode.js (存在しない場合) が自動的にインストールされ、`config.toml` がセットアップされ、Vue 3フロントエンドがビルドされ、Goバックエンドがコンパイルされます。

1.  **リポジトリのクローン:**
    ```bash
    git clone https://github.com/your-org/polytrade-bot.git
    cd polytrade-bot
    ```

2.  **セットアップスクリプトを実行する:**
    ```bash
    ./setup.sh
    ```

#### オプション 2: 手動セットアップ
1.  **リポジトリのクローン:**
    ```bash
    git clone https://github.com/your-org/polytrade-bot.git
    cd polytrade-bot
    ```

2.  **ボットの設定:**
    ルートディレクトリに`config.toml`ファイルを作成します。ファイルなしでボットを起動し、TUIウィザードを使用して`private_key`を安全に設定することもできます。

3.  **ビルドと実行:**
    ```bash
    # バイナリのビルド
    go build ./...

    # ボットの実行
    go run ./cmd/bot/ --config config.toml
    ```

### ヘッドレスモード
TUIを持たないサーバー環境でボットを実行する場合は、以下のフラグを使用します：
```bash
go run ./cmd/bot/ --config config.toml --no-tui
```

## 🛠 トラブルシューティングとよくある問題

*   **API キー / 401 Unauthorized:** `private_key`が正しいことを確認してください。ボットは起動時にL2 APIキーを自動的に取得します。L2署名は30秒で期限切れになるため、サーバーの時刻が同期されていること（NTP）を確認してください。
*   **Web UI "Network Error":** Go HTTPハンドラーでパニックが発生した場合、GoはJSON本文なしでTCP接続を閉じるため、ブラウザは一般的な"Network Error"を報告します。ターミナルログで実際のパニックスタックトレースを確認してください。
*   **UIで市場データが欠落している:** イベントバス（EventBus）は、バッファがいっぱいになるとメッセージを静かに破棄します。ログレベルが`trace`に設定されていてログが多すぎる場合、`MarketsUpdatedMsg`などの重要なメッセージが破棄される可能性があります。ログレベルを`info`または`debug`に下げてください。
*   **Polymarket Token IDs のパース:** Gamma APIのトークンIDは10進数の文字列です。16進数としてパースしようとしないでください。署名が無効になります。ボットは`big.Int.SetString(id, 10)`を使用して正しく処理します。

## 💻 開発ガイド

### Web UIのビルド
Vue 3 Web UIはGoバイナリに組み込まれています。`internal/webui/web/src`内のファイルを変更した場合は、変更を適用するためにフロントエンドを再ビルドする必要があります：
```bash
cd internal/webui/web
npm install
npm run build
```

### ボットの拡張
*   **新しい取引戦略:** `trading.Strategy`インターフェース（`Name`、`Start`、`Stop`）を実装し、`main.go`でインスタンス化して、`engine.Register(s)`を呼び出します。
*   **新しい設定項目の追加:** `tab_settings.go`、`Locale`構造体にフィールドを追加し、5つの`locales/*.json`ファイルを更新して、`config_key.go`の`applyConfigKey()`にロジックを追加します。`SettingsView.vue`も更新してください。
*   **新しいTelegramコマンド:** `internal/telegrambot/handlers.go`の`handleCommand` switchステートメント内にハンドラーを追加します。

### テストの実行
```bash
# ユニットテスト
go test ./...

# 統合テスト (実際のPolymarket APIとL1キーが必要)
POLY_PRIVATE_KEY=0xあなたのキー go test ./... -tags=integration -timeout 90s
```
