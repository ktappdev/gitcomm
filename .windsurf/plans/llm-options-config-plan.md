# Move LLM options to config
This plan moves all LLM settings into the config file (as the primary source) while keeping hardcoded defaults as fallbacks.

## Proposed changes
- Extend `internal/config/config.go` with LLM settings: `models`, `max_tokens`, `temperature`, `api_url`, `timeout_seconds`, and keep `open_router_api_key`.
- Ensure load logic applies defaults when values are missing or invalid, with environment variable overriding only the API key.
- Update `internal/llm/client.go` to use config-provided values first, then fall back to the existing hardcoded defaults.
- Provide an example config: add JSON example in README and add a sample config file in the repo (non-secret).
- Update README to document new config fields, defaults, and precedence.

## Open questions
- Where should the example config live in the repo (e.g., `config.example.json` at root, or `docs/config.example.json`)?
- Should the setup wizard only ask for the API key (current behavior), or also prompt for LLM settings?
