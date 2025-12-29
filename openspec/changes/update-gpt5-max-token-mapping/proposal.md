# Change: Update GPT-5 max_tokens normalization

## Why
GPT-5 series endpoints reject the legacy `max_tokens` parameter, causing upstream errors when clients send older payloads.

## What Changes
- Normalize GPT-5* requests by translating legacy `max_tokens` into `max_completion_tokens` before forwarding.
- Drop unsupported `max_tokens` when both fields are present while preserving the provided completion limit.
- Add regression coverage to ensure GPT-5 requests are forwarded without unsupported parameters.

## Impact
- Affected specs: `openai-compatibility` (new capability)
- Affected code: OpenAI adaptor request conversion logic and related tests for GPT-5 parameter normalization
