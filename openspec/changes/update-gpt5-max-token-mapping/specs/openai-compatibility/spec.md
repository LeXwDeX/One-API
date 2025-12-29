## ADDED Requirements
### Requirement: GPT-5 max_tokens normalization
System SHALL normalize GPT-5 series requests by converting legacy `max_tokens` into `max_completion_tokens` before sending to upstream OpenAI-compatible APIs.

#### Scenario: Legacy max_tokens only
- **WHEN** a request targets a GPT-5* model and includes `max_tokens` but not `max_completion_tokens`
- **THEN** the system maps that value to `max_completion_tokens` and omits `max_tokens` in the forwarded payload

#### Scenario: Both max fields provided
- **WHEN** a GPT-5* request contains both `max_tokens` and `max_completion_tokens`
- **THEN** the system preserves `max_completion_tokens`, removes `max_tokens`, and forwards only the supported field
