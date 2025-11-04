# Trading agent backend

- Gin server
- PostgreSQL via sqlx
- Structured logging via zap
- Simple schema auto-migration on start
- Services:
  - wallet: connect and store encrypted API key (AES-GCM using `SECRET_KEY`)
  - bot: start/stop background loop; in sandbox generates sample trades and updates stats
  - stats: latest balance, pnl, roe
  - trades: insert/list
- Endpoints:
  - POST `/api/wallet/connect`
  - POST `/api/bot/start`
  - POST `/api/bot/stop`
  - GET `/api/stats`
  - GET `/api/trades/history?limit=100`
  - Swagger UI: GET `/swagger` (spec at `/swagger/openapi.json`)

Live HyperLiquid client is used; configure API secrets in environment.

### DeepSeek-driven Trading Bot

- Backend integrates a DeepSeek agent to decide actions (buy/sell/none) using chat-completions JSON output.
- Configure:
  - `DEEPSEEK_API_KEY` (required to enable agent)
  - `DEEPSEEK_BASE_URL` (default `https://api.deepseek.com`)
  - `DEEPSEEK_MODEL` (default `deepseek-chat`)
- Bot periodically builds a snapshot (live balance/pnl/roe + recent trades), asks the agent, and places orders via HyperLiquid client (when wallet is connected).
- Inspired by agent-driven design and reporting in AI-Trader. See: `https://github.com/HKUDS/AI-Trader`

### Live mode (real signing and orders)

- Set `SANDBOX=false` and provide your agent wallet private key in `API_SECRET` (hex). `API_KEY` can store the public address (optional for reference).
- `HL_BASE_URL` and `HL_WS_URL` can be overridden; defaults use mainnet endpoints.
- The live client uses the Go SDK to sign with secp256k1 and submit orders, per the official docs.

Docs and references:

- HyperLiquid API: `https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api`
- Go SDK used: `https://github.com/sonirico/go-hyperliquid`
