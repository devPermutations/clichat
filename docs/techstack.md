## Tech Stack

## Language and Runtime
- **Go**: >= 1.22

## CLI Framework
- **spf13/cobra** for commands. Shell completion not implemented yet.
- In-session line editing + tab completion: `github.com/peterh/liner` (cross-platform, lightweight).

## LLM Provider
- **LiteLLM proxy** as the initial provider.
- Requirements: token streaming, temperature/top_p control, model selection.
- Base URL and API key provided via .env.

## Streaming
- Use LiteLLM's streaming (SSE/chunked) with incremental stdout rendering.

## Configuration
- **.env** file for keys and defaults using `github.com/joho/godotenv`.
- Env-only for now (no flags). Keys include: `LITELLM_BASE_URL`, `LITELLM_API_KEY`, `LLM_MODEL`, `TEMPERATURE`, `TOP_P`, `DB_PATH`, `SYSTEM_PROMPT`, `MODEL_CONTEXT_TOKENS`, `ENABLE_PROVIDER_WEBSEARCH`, `DROP_SAMPLING_PARAMS`, `DEBUG_PROMPTS`.

## Persistence / Memory
- **SQLite** for local history.
- Preferred driver (no CGO): `modernc.org/sqlite`.
- Access via `database/sql` (no heavy ORM).

## Logging
- Using fmt/stdout-stderr for now; can migrate to `log/slog` later.

## Tools / Function Calling
- Pass tool definitions to compatible models via LiteLLM (OpenAI-style tools).
 - MVP: provider-native only. If the selected model exposes browsing via LiteLLM (e.g., `web_search`), enable it on the provider/LiteLLM side; the CLI passes through `tools` and does not implement local tool execution.

## Testing
- `go test` with table-driven tests; avoid network calls (mock provider).
- Mock tool execution; verify tool-call handling paths.

## Packaging
- Plain `go build`; optional `goreleaser` later.
- Docker: multi-stage build; CGO-disabled static binary for Linux container.

## Linting/Formatting
- `gofmt`, `go vet`; optional `golangci-lint` later.

## Token Accounting
- Local estimation based on rough chars-per-token heuristic; context percent printed when `MODEL_CONTEXT_TOKENS` is set.


