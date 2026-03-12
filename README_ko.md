# Polytrade Bot 📈🤖

*다른 언어로 읽기: [English](README.md), [Русский](README_ru.md), [中文](README_zh.md), [日本語](README_ja.md).*

Polytrade Bot은 **Polymarket CTF Exchange**를 위한 고급 알고리즘 트레이딩 및 관리 봇입니다. 인터랙티브 터미널 사용자 인터페이스(TUI), Vue 3 Web UI 및 원격 관리를 위한 Telegram 봇을 포함한 강력한 다중 인터페이스 아키텍처를 특징으로 합니다.

## 🌟 상세 기능

*   **다중 인터페이스 경험:**
    *   **TUI:** 시장, 거래, 카피트레이딩, 지갑, 전략, 설정 및 로그를 위한 개별 탭이 있는 아름다운 터미널 인터페이스.
    *   **Web UI:** 실시간 WebSocket 업데이트, JWT 인증 및 반응형 디자인을 갖춘 Vue 3 SPA.
    *   **Telegram Bot:** TUI 기능을 완벽하게 미러링하며, 인라인 키보드 및 다단계 대화를 지원하는 인터랙티브 봇.
*   **알고리즘 트레이딩 엔진:** 차익 거래(Arbitrage), 크로스 마켓, Fade Chaos, 마켓 메이킹 등의 전략 내장. 사용자 지정 전략을 쉽게 등록할 수 있습니다.
*   **고급 카피 트레이딩:** Data API를 통해 타겟 지갑을 실시간으로 모니터링하고 CLOB API를 통해 자동으로 포지션을 복사합니다. 동적 볼륨 할당 모드(`proportional` 또는 `fixed_pct`)를 지원합니다.
*   **실시간 모니터링 및 알림:**
    *   **거래 모니터:** 미체결 주문, 거래 체결 및 포지션을 추적합니다.
    *   **시장 알림:** 시장 상태 변화에 따른 알림 조건을 실시간으로 평가합니다.
*   **강력한 보안 인증:** L1/L2 자격 증명 아키텍처. EIP-712 서명을 자동으로 파생하며, L2 키는 메모리에만 유지되고 구성 파일에 절대 저장되지 않습니다. 보안을 위해 서명은 30초 후에 자동 만료됩니다.
*   **다중 지갑 지원:** 인터페이스에서 직접 여러 활성 지갑을 관리하고, 켜거나 끄며 집계된 통계를 확인할 수 있습니다.
*   **다국어 지원 (i18n):** 영어, 러시아어, 중국어, 일본어, 한국어를 기본으로 지원하며, 재시작 없이 즉시 언어 변경이 가능합니다.

## 🏗 아키텍처 개요

이 봇은 7개의 핵심 컨텍스트 취소 가능 하위 시스템으로 작동합니다:

1.  **WebSocket 클라이언트:** Polymarket CLOB(`market`, `user`, `asset` 채널)에 대한 자동 재연결이 가능한 영구 연결.
2.  **모니터:** Gamma 및 Data API를 폴링하여 시장 상태 변화를 평가하고 알림을 생성합니다.
3.  **트레이딩 엔진:** 플러그형 트레이딩 전략(`trading.Strategy`) 실행을 위한 확장 가능한 Goroutine 기반 계층.
4.  **알림 시스템:** 구성 가능한 알림 시스템(기본값: Telegram).
5.  **카피 트레이더:** 설정된 지갑을 추적하고 포지션을 복제합니다. 봇을 다시 시작하지 않아도 구성 변경 시 핫 리로드를 지원합니다.
6.  **Telegram Bot:** 단일 관리자 모델(`AdminChatID`)을 사용하는 TUI의 대화형 미러.
7.  **Web UI:** Vue 3 SPA를 제공하는 내장 HTTP 서버 + WebSocket 허브.

## ⚙️ 구성 (`config.toml`)

봇은 완전히 `config.toml`에 의해 제어됩니다. 안전을 위해 트레이딩 및 데이터베이스 기능은 기본적으로 비활성화되어 있습니다.

주요 섹션:
*   `[auth]`: `private_key`(16진수, `0x` 접두사 없음)가 필요합니다. L2 자격 증명은 시작 시 자동으로 파생됩니다.
*   `[webui]`: `enabled` (true/false), `listen` (예: `127.0.0.1:8080`), `jwt_secret` (서명 및 로그인 암호로 사용됨).
*   `[ui]`: `language` (`en`, `ru`, `zh`, `ja`, `ko`). 즉시 핫 리로드됩니다.
*   `[monitor.trades]`: `enabled`, `poll_interval_ms`. L2 인증이 필요합니다.
*   `[copytrading]`: `enabled`, `size_mode` (`proportional`/`fixed_pct`) 및 `[[copytrading.traders]]` 목록. 데이터베이스 및 L2가 필요합니다.
*   `[telegram]`: `enabled`, `bot_token`, `admin_chat_id` (단일 관리자 대상).
*   `[database]`: `enabled`, `path` (SQLite DB 경로).
*   `chain_id`: Polygon 메인넷의 경우 `137`, Amoy 테스트넷의 경우 `80002`.

## 🚀 설치 및 설정

### 사전 요구 사항

*   [Go 1.24+](https://golang.org/doc/install)
*   [Node.js 18+](https://nodejs.org/) (Web UI를 수정할 때만 필요)
*   Polymarket 지갑 개인 키

### 설치 단계

#### 옵션 1: 범용 설정 스크립트 (권장)
Linux, macOS 및 Windows(Git Bash/WSL 경유)에서 작동하는 범용 `setup.sh` 스크립트를 제공합니다. Go 및 Node.js(없는 경우)를 자동으로 설치하고, `config.toml`을 설정하고, Vue 3 프론트엔드를 빌드하고, Go 백엔드를 컴파일합니다.

1.  **저장소 클론:**
    ```bash
    git clone https://github.com/your-org/polytrade-bot.git
    cd polytrade-bot
    ```

2.  **설정 스크립트 실행:**
    ```bash
    ./setup.sh
    ```

#### 옵션 2: 수동 설정
1.  **저장소 클론:**
    ```bash
    git clone https://github.com/your-org/polytrade-bot.git
    cd polytrade-bot
    ```

2.  **봇 설정:**
    루트 디렉토리에 `config.toml` 파일을 생성합니다. 파일 없이 봇을 시작하여 TUI 마법사를 통해 `private_key`를 안전하게 구성할 수도 있습니다.

3.  **빌드 및 실행:**
    ```bash
    # 바이너리 빌드
    go build ./...

    # 봇 실행
    go run ./cmd/bot/ --config config.toml
    ```

### 헤드리스 모드 (Headless Mode)
TUI가 없는 서버 환경에서 봇을 실행하려면 다음 플래그를 사용하세요:
```bash
go run ./cmd/bot/ --config config.toml --no-tui
```

## 🛠 문제 해결 및 일반적인 문제

*   **API Key / 401 Unauthorized:** `private_key`가 올바른지 확인하세요. 봇은 시작 시 L2 API 키를 자동으로 파생합니다. L2 서명은 30초 후에 만료되므로 서버 시간이 동기화되어 있는지(NTP) 확인하세요.
*   **Web UI "Network Error":** Go HTTP 핸들러에서 패닉(panic)이 발생하면 Go가 JSON 본문 없이 TCP 연결을 닫기 때문에 브라우저에서 일반적인 "Network Error"를 보고합니다. 터미널 로그에서 실제 Go 패닉 스택 추적을 확인하세요.
*   **UI에서 시장 데이터 누락:** 내부 EventBus는 버퍼가 가득 차면 메시지를 조용히 삭제합니다. 로그 수준이 `trace`로 설정되어 있고 로그가 너무 많이 생성되면 `MarketsUpdatedMsg`와 같은 중요한 메시지가 삭제될 수 있습니다. 로그 수준을 `info` 또는 `debug`로 낮추세요.
*   **Polymarket Token IDs 파싱:** Gamma API의 토큰 ID는 10진수 문자열입니다. 이를 직접 16진수로 파싱하려고 시도하지 마세요. 서명이 무효화됩니다. 봇은 `big.Int.SetString(id, 10)`을 사용하여 올바르게 처리합니다.

## 💻 개발 가이드

### Web UI 빌드
Vue 3 Web UI는 Go 바이너리에 내장되어 있습니다. `internal/webui/web/src`의 파일을 수정한 경우 변경 사항을 적용하려면 프론트엔드를 다시 빌드해야 합니다:
```bash
cd internal/webui/web
npm install
npm run build
```

### 봇 기능 확장
*   **새로운 트레이딩 전략:** `trading.Strategy` 인터페이스(`Name`, `Start`, `Stop`)를 구현하고, `main.go`에서 인스턴스화한 후 `engine.Register(s)`를 호출합니다.
*   **새로운 설정 항목 추가:** `tab_settings.go`, `Locale` 구조체에 필드를 추가하고 5개의 `locales/*.json` 파일을 업데이트한 후 `config_key.go`의 `applyConfigKey()`에 논리를 추가합니다. `SettingsView.vue`도 업데이트하세요.
*   **새로운 Telegram 명령어:** `internal/telegrambot/handlers.go`의 `handleCommand` switch 문 내에 핸들러를 추가합니다.

### 테스트 실행
```bash
# 단위 테스트
go test ./...

# 통합 테스트 (실제 Polymarket API 및 L1 키 필요)
POLY_PRIVATE_KEY=0x당신의_키 go test ./... -tags=integration -timeout 90s
```
