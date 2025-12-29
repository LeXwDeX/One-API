## ADDED Requirements
### Requirement: Webhook Configuration
System SHALL allow admins to configure HTTPS webhook endpoints for usage/billing events with URL, shared secret, enabled flag, subscribed event types, and a test-delivery check.

#### Scenario: Admin saves a webhook
- **WHEN** an admin provides a valid HTTPS URL, secret, and selected event types
- **THEN** the system validates reachability with a test request and persists the configuration as enabled

### Requirement: Usage Event Delivery
System SHALL emit usage events (success, failure, quota/rate-limit denial) asynchronously without blocking user requests, including request identifiers, token/channel/user, model, token usage, cost multipliers, and latency.

#### Scenario: Successful request emits an event
- **WHEN** a relay request completes successfully
- **THEN** the system enqueues a usage event payload for delivery to all enabled webhooks subscribed to success events

#### Scenario: Failure emits an event
- **WHEN** a relay request fails due to upstream error or quota/rate limit
- **THEN** the system enqueues a failure payload with status, error code/message, and consumed quota (if any)

### Requirement: Delivery Reliability and Security
Webhook deliveries SHALL be signed (e.g., HMAC of body with shared secret), retried with exponential backoff on failure, and surfaced with status/history for operators.

#### Scenario: Delivery is retried and reported
- **WHEN** a webhook endpoint returns a non-2xx status or times out
- **THEN** the system retries delivery up to the configured attempts with backoff, records attempts/results, and marks the webhook as failing after exhausting retries while leaving user traffic unaffected
