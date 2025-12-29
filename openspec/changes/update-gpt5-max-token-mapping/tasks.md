## 1. Implementation
- [x] 1.1 Detect GPT-5-prefixed models in the OpenAI adaptor and rewrite legacy `max_tokens` into `max_completion_tokens` while omitting the unsupported field.
- [x] 1.2 Handle requests that include both fields by preferring `max_completion_tokens`, stripping `max_tokens`, and forwarding the normalized payload.
- [x] 1.3 Add regression tests that cover GPT-5 requests with legacy/both fields to verify only `max_completion_tokens` is sent and update any relevant docs or logs if needed.
