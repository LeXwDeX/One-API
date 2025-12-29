# Change: Add usage webhooks for request and billing events

## Why
Teams integrating One API want to push usage and failure events into their own billing, alerting, or observability stacks without polling the admin UI or database.

## What Changes
- Add admin-facing configuration for HTTPS webhook endpoints with shared secret, event filtering, and a test-delivery check.
- Emit usage/billing events (success, error, quota exhaustion) asynchronously with signed payloads and retry/backoff.
- Surface delivery status and failures in the admin UI/monitor logs without blocking user traffic.

## Impact
- Affected specs: `usage-webhooks` (new capability)
- Affected code: admin config APIs/UI, usage logging/monitor, background delivery worker, configuration storage/migrations, docs
