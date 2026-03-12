# Polytrade Bot 📈🤖

*阅读其他语言版本: [English](README.md), [Русский](README_ru.md), [한국어](README_ko.md), [日本語](README_ja.md).*

Polytrade Bot 是一个专为 **Polymarket CTF 交易所**设计的高级算法交易与管理机器人。它拥有强大的多终端架构，包括交互式终端用户界面 (TUI)、Vue 3 Web UI 以及用于远程管理的综合 Telegram 机器人。

## 🌟 详细功能

*   **多终端体验：**
    *   **TUI:** 美观的终端界面，包含市场、交易、跟单、钱包、策略、设置和日志等独立选项卡。
    *   **Web UI:** 基于 Vue 3 的 SPA，支持 WebSocket 实时更新、JWT 身份验证以及响应式设计。
    *   **Telegram Bot:** 具有内联键盘和多步对话功能的交互式机器人，完美映射 TUI 的核心功能。
*   **算法交易引擎：** 内置套利、跨市场、Fade Chaos、做市等多种策略。支持轻松注册和接入自定义交易策略。
*   **高级跟单交易 (Copy Trading)：** 通过 Data API 实时监控目标钱包，并通过 CLOB API 自动复制头寸。支持动态仓位分配模式（`proportional` 按比例或 `fixed_pct` 固定百分比）。
*   **实时监控与警报：**
    *   **交易监控:** 跟踪未结订单、交易执行和持仓情况。
    *   **市场警报:** 根据市场状态的变化实时评估警报条件并推送通知。
*   **安全认证：** 采用 L1/L2 凭证架构。自动派生 EIP-712 签名——您的 L2 密钥完全在内存中生成，绝不会被写入配置文件。签名将在 30 秒后自动过期，以确保极致的安全。
*   **多钱包支持：** 直接从 UI 界面轻松管理多个活跃钱包，快速开启/关闭并查看汇总统计信息。
*   **国际化 (i18n)：** 所有界面原生支持英语、俄语、中文、日语和韩语，并支持热重载（即时切换）。

## 🏗 架构概述

该机器人基于七个核心的可取消上下文子系统运行：

1.  **WebSocket 客户端：** 与 Polymarket CLOB (`market`, `user`, `asset` 频道) 保持持久连接并支持自动重连。
2.  **监控器：** 轮询 Gamma & Data API，评估市场差异并触发警报。
3.  **交易引擎：** 基于 Goroutine 的可扩展执行层，用于运行可插拔的交易策略 (`trading.Strategy`)。
4.  **通知系统：** 可配置的警报系统（默认发送至 Telegram）。
5.  **跟单交易：** 跟踪配置的钱包并复制头寸。支持在不重启机器人的情况下热重载配置更改。
6.  **Telegram Bot：** 使用单管理员模式 (`AdminChatID`) 的交互式 TUI 镜像。
7.  **Web UI：** 嵌入式 HTTP 服务器 + WebSocket Hub，用于提供 Vue 3 SPA 服务。

## ⚙️ 配置文件 (`config.toml`)

机器人完全由 `config.toml` 控制。为了安全起见，交易和数据库功能默认处于禁用状态。

关键配置项：
*   `[auth]`: 需要 `private_key` (十六进制，无 `0x` 前缀)。L2 凭据在启动时自动获取。
*   `[webui]`: `enabled` (true/false), `listen` (例如 `127.0.0.1:8080`), `jwt_secret` (用于签名和作为 Web 登录密码)。
*   `[ui]`: `language` (`en`, `ru`, `zh`, `ja`, `ko`)。支持即时生效。
*   `[monitor.trades]`: `enabled`, `poll_interval_ms`。需要 L2 认证。
*   `[copytrading]`: `enabled`, `size_mode` (`proportional`/`fixed_pct`)，以及 `[[copytrading.traders]]` 列表。需要数据库和 L2。
*   `[telegram]`: `enabled`, `bot_token`, `admin_chat_id` (唯一的管理员聊天 ID)。
*   `[database]`: `enabled`, `path` (SQLite 数据库路径)。
*   `chain_id`: Polygon 主网为 `137`，Amoy 测试网为 `80002`。

## 🚀 安装与设置

### 环境要求

*   [Go 1.24+](https://golang.org/doc/install)
*   [Node.js 18+](https://nodejs.org/) (仅在修改 Web UI 时需要)
*   Polymarket 钱包私钥

### 安装步骤

#### 选项 1：通用安装脚本 (推荐)
我们提供了一个通用的 `setup.sh` 脚本，适用于 Linux、macOS 和 Windows (通过 Git Bash/WSL)。它会自动安装 Go 和 Node.js (如果缺失)，设置您的 `config.toml`，构建 Vue 3 前端，并编译 Go 后端。

1.  **克隆仓库：**
    ```bash
    git clone https://github.com/your-org/polytrade-bot.git
    cd polytrade-bot
    ```

2.  **运行安装脚本：**
    ```bash
    ./setup.sh
    ```

#### 选项 2：手动设置
1.  **克隆仓库：**
    ```bash
    git clone https://github.com/your-org/polytrade-bot.git
    cd polytrade-bot
    ```

2.  **配置机器人：**
    在根目录创建 `config.toml` 文件。您也可以直接启动机器人，TUI 设置向导将引导您安全地输入 `private_key`。

3.  **构建与运行：**
    ```bash
    # 编译二进制文件
    go build ./...

    # 运行机器人
    go run ./cmd/bot/ --config config.toml
    ```

### 无头模式 (Headless Mode)
如果要在没有 TUI 的服务器环境中运行，请使用 `--no-tui` 参数：
```bash
go run ./cmd/bot/ --config config.toml --no-tui
```

## 🛠 常见问题排查

*   **API 密钥 / 401 Unauthorized:** 确保您的 `private_key` 正确。机器人会自动获取 L2 API 密钥。L2 签名在 30 秒后过期；请确保您的服务器时间已同步 (NTP)。
*   **Web UI "Network Error":** 如果 Go HTTP 处理程序发生 panic，浏览器会提示通用的 "Network Error"，因为 Go 在没有返回 JSON 响应体的情况下关闭了 TCP 连接。请在终端日志中检查实际的 panic 堆栈。
*   **UI 中丢失市场数据:** 如果事件总线 (EventBus) 的缓冲区已满，它会静默丢弃消息。如果日志级别设置为 `trace`，产生的大量日志可能会挤掉如 `MarketsUpdatedMsg` 这样的重要消息。建议将日志级别降低到 `info` 或 `debug`。
*   **Polymarket Token IDs 解析:** Gamma API 返回的 Token ID 是十进制字符串。请勿将其作为十六进制进行解析，否则会导致签名无效。机器人在内部使用 `big.Int.SetString(id, 10)` 进行正确处理。

## 💻 开发指南

### 构建 Web UI
Web UI 嵌入在 Go 二进制文件中。如果修改了 `internal/webui/web/src` 中的文件，必须重新构建前端：
```bash
cd internal/webui/web
npm install
npm run build
```

### 扩展机器人功能
*   **新增交易策略:** 实现 `trading.Strategy` 接口 (`Name`, `Start`, `Stop`)，在 `main.go` 中实例化，并调用 `engine.Register(s)`。
*   **新增配置项:** 在 `tab_settings.go` 和 `Locale` 结构体中添加字段，更新 5 个 `locales/*.json` 文件，并在 `config_key.go` 的 `applyConfigKey()` 中添加逻辑。同时更新 `SettingsView.vue` 及其语言文件。
*   **新增 Telegram 命令:** 在 `internal/telegrambot/handlers.go` 的 `handleCommand` switch 语句中添加处理程序。

### 运行测试
```bash
# 单元测试
go test ./...

# 集成测试 (需要真实的 Polymarket API 和 L1 私钥)
POLY_PRIVATE_KEY=0x您的私钥 go test ./... -tags=integration -timeout 90s
```
