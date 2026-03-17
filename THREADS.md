# X.com Posts / Threads — Orbitron

---

## Post 1 — Launch Announcement (Main Thread)

**Tweet 1/6**
```
Introducing Orbitron — an open-source algorithmic trading & portfolio management bot for @Polymarket

Self-hosted. Multi-interface. Built for serious prediction market traders.

github.com/atlas-is-coding/orbitron-polymarket-system
getorbitron.net

[Alpha — bugs expected. Details below]
```

**Tweet 2/6**
```
Three ways to control it:

- Terminal UI (BubbleTea) — tabbed dashboard, live P&L, strategy controls
- Web UI (Vue 3) — real-time WebSocket, JWT auth, dark theme
- Telegram Bot — inline keyboards, alerts, full remote control

One binary. All three run simultaneously.
```

**Tweet 3/6**
```
6 built-in trading strategies, all pluggable:

- Arbitrage — YES/NO spread capture
- Cross-Market — correlated divergence
- Fade Chaos — mean reversion after spikes
- Market Making — resting limit orders
- Positive EV — mispriced probability scanner
- Riskless Rate — near-resolved binary markets

Add your own in 3 lines of Go.
```

**Tweet 4/6**
```
Security-first design:

- L1 EIP-712 + L2 HMAC-SHA256 auth
- API keys derived in-memory at startup, never written to disk
- Multi-wallet support with per-wallet toggle
- Trading disabled by default — you opt in

Built on Polygon. Settles in USDC.e.
```

**Tweet 5/6**
```
Also ships with:

- Copy trading — mirror target wallets via Data API + CLOB
- Price alerts & real-time fill notifications to Telegram
- 5 UI languages (EN/RU/ZH/JA/KO) with hot-reload
- SQLite persistence
- HTTP/SOCKS5 proxy support

One config.toml to rule them all.
```

**Tweet 6/6**
```
This is an ALPHA release.

Bugs will happen. Trading on real funds at your own risk.

Found something broken?
- Open an issue: github.com/atlas-is-coding/orbitron-polymarket-system/issues
- Email: mw.atlas.kun@gmail.com
- DM me on X

Stars and feedback appreciated.

getorbitron.net
```

---

## Post 2 — Short Promo (Single Tweet)

```
Built an open-source algo trading bot for @Polymarket prediction markets.

6 strategies. Terminal UI + Web UI + Telegram bot. Copy trading. Multi-wallet. Self-hosted.

Alpha release — expect rough edges, report bugs.

getorbitron.net
github.com/atlas-is-coding/orbitron-polymarket-system
```

---

## Post 3 — Technical / Builder Audience

```
If you trade on @Polymarket and know Go:

Orbitron is an open-source bot with a pluggable strategy interface.

Implement 3 methods, register one line — your strategy runs inside a live trading engine with L2 HMAC auth, orderbook access, and real-time WebSocket feeds.

Alpha. MIT. github.com/atlas-is-coding/orbitron-polymarket-system
```

---

## Post 4 — Copy Trading Focus

```
Copy trading on @Polymarket is now open-source.

Orbitron monitors any wallet via the Data API and auto-replicates their positions through the CLOB — proportional or fixed %.

Alpha release. Self-hosted. Free.

getorbitron.net
```

---

## Form Fields (image — Builder/Grant Application)

| Field | Value |
|---|---|
| **Product Name** | Orbitron |
| **Project Description** | Orbitron is an open-source, self-hosted algorithmic trading and portfolio management bot for the Polymarket CTF Exchange. It combines a pluggable 6-strategy trading engine, copy trading, real-time market monitoring, and a multi-wallet architecture — all controllable from a Terminal UI (BubbleTea), Web UI (Vue 3), or Telegram bot. Built in Go, licensed MIT. |
| **Website URL** | https://getorbitron.net |
| **Email** | mw.atlas.kun@gmail.com |
| **X Handle** | *(your @handle)* |
| **Telegram Handle** | *(your @handle)* |
| **Builder API key** | *(leave blank or add if you have one)* |
