# CRUSH.md

Build/lint/test
- Build: go build ./...
- Run: go run .
- Install CLI: go install ./...
- Tidy deps: go mod tidy
- Lint: golangci-lint run (if not installed: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
- Format: go fmt ./...
- Vet: go vet ./...
- Tests (all): go test ./...
- Tests (package): go test ./internal/analyzer -v
- Tests (single): go test ./internal/analyzer -run TestName -v

Project conventions (Go)
- Go version: from go.mod (go 1.23.x). Modules enabled.
- Imports: stdlib first, then external, then internal (github.com/ktappdev/gitcomm/internal/...). Use goimports.
- Formatting: go fmt; no comments unless necessary. Keep files under 250 lines when practical.
- Types: prefer explicit types; return (T, error). Avoid panics in library code.
- Errors: wrap with context using fmt.Errorf("...: %w", err); never log secrets; user-facing CLI errors printed with fmt.Println in main.
- Logging: use log/slog where structured logs are needed (already used in internal/llm).
- Naming: exported identifiers use CamelCase; unexported lowerCamel; package names short and lower case.
- Config: API keys via env (GEMINI_API_KEY, GROQ_API_KEY, OPENAI_API_KEY) or ~/.gitcomm/config.json; env overrides file.
- LLM: default provider OpenRouter; models set in internal/llm; 400 tokens for proper Git format (subject + body).
- Git ops: internal/git shells out to git; avoid interactive prompts in library code.
- CLI flags: -setup, -auto, -ap, -sa; keep backwards-compatible.

Assistant rules
- If .cursor/rules/ or .cursorrules or .github/copilot-instructions.md appear, mirror their guidance here; none found currently.
- Add frequently used commands here when discovered; prefer smallest reproducible commands.
