# Project Context

## Purpose
Provide a unified, OpenAI-compatible API gateway for multiple LLM providers with routing, load balancing, quota/usage controls, and an admin UI so teams can manage access keys, channels, and billing-like quotas in one place.

## Tech Stack
- Backend: Go 1.20, Gin, GORM (MySQL/PostgreSQL/SQLite), optional Redis cache, JWT/sessions, WebSocket support, Docker/Compose & systemd for deployment.
- Frontend: React 18 (CRA), Semantic UI React, Recharts, i18next, axios; theming via multiple `web/*` themes with Prettier single-quote config.
- Tooling: Go modules, `gofmt`, npm scripts for frontend (`build` outputs to `web/build/<theme>`), optional Message Pusher/Turnstile integrations.

## Project Conventions

### Code Style
- Go code is formatted with `gofmt`; follow idiomatic Go naming and keep handler/service logic small and readable.
- Web code follows CRA defaults with Prettier single quotes; keep components functional and colocate translations/assets with pages.
- Configuration is environment-driven (`.env` supported via `godotenv`); prefer centralized constants/config structs.

### Architecture Patterns
- Monolithic Go service: `router` + `controller` for HTTP, `middleware` for auth/rate limit/logging, `common` for config/util, `model` for persistence via GORM, `relay` for provider adapters, `monitor` for background jobs.
- Static frontend assets are built under `web/<theme>/build` and served by the Go server; theming is registered in `common/config/config.go` and `web/THEMES`.
- Multi-tenant abstractions: accounts, tokens, channels, channel/user groups, model mappings, and provider-specific relay logic; supports master/slave node roles with optional Redis sync.

### Testing Strategy
- Unit tests via `go test` with Testify/Goconvey for core logic; prefer table-driven tests around controllers, relay adapters, and billing/quota calculations.
- Frontend smoke tests available via CRA `npm test` (Jest); prioritize critical flows (auth, token creation, channel operations) when modifying UI logic.
- Manual verification for provider integrations, streaming, and multi-node/cache behaviors; prefer adding regression tests for quota and rate-limit paths.

### Git Workflow
- Default branch is `main`; use short-lived feature branches and PRs for changes.
- Keep commits small and descriptive; avoid rewriting shared history; align releases with `VERSION` tagging.

## Domain Context
- Core domain: relay OpenAI-format requests to many upstream LLM providers (OpenAI/Azure, Claude, Gemini, Mistral, Baidu, Alibaba, Tencent, Groq, DeepSeek, etc.) with load balancing, retries, and per-channel model allowlists.
- Admin UX covers token issuance with quotas/expiry/IP allowlists, channel management, model mapping, coupons, user/group pricing multipliers, announcements, and invite rewards.
- Supports streaming responses, drawing/image endpoints, Cloudflare Turnstile checks, OAuth (GitHub/Feishu/WeChat) logins, and theming.

## Important Constraints
- Must remain API-compatible with OpenAI schemas (including streaming) while normalizing provider-specific differences; avoid breaking default admin login/bootstrap.
- Multi-database support (SQLite for dev, MySQL/PostgreSQL for production) must stay intact; Redis is optional and should not be assumed.
- Concurrency-focused deployment: respect `SESSION_SECRET`, `NODE_TYPE`, cache sync intervals, and rate-limit settings; preserve MIT license attribution in UI unless separately licensed.
- Compliance: features must respect upstream providers’ ToS and local regulations for generative AI exposure.

## External Dependencies
- Datastores: MySQL/PostgreSQL/SQLite; optional Redis for caching and sync.
- Upstream providers: OpenAI/Azure and other LLM vendors listed in README (Claude, Gemini, Mistral, Baidu/Wenxin, Qwen, Spark, Zhipu, Hunyuan, Groq, DeepSeek, etc.).
- Integrations: Cloudflare Turnstile, Message Pusher for alerts, SMTP/email for password reset, OAuth providers (GitHub/Feishu/WeChat), Nginx/HTTPS termination for deployment.
