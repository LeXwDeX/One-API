## 1. Implementation
- [ ] 1.1 Add data model + migration for webhook configs (URL, secret, enabled flag, subscribed events, retry policy, last status).
- [ ] 1.2 Expose admin API + UI to create/update/disable/test webhook endpoints with validation and sample payload preview.
- [ ] 1.3 Emit usage events (success, error, quota/rate-limit) into an async queue without blocking request handling.
- [ ] 1.4 Implement delivery worker with HMAC signing, retry/backoff, DLQ/failure logging, and metrics.
- [ ] 1.5 Add tests for payload shape, signing, retries, and admin validation; update docs and environment variable references.
