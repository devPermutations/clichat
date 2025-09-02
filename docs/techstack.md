## Tech Stack

## Language and Runtime
- **Go**: >= 1.22

## CLI Framework
- **spf13/cobra** for commands and shell completion (bash/zsh/fish/powershell).
- In-session line editing + tab completion: `github.com/peterh/liner` (cross-platform, lightweight).
 - Ship bash completion first; add others later.

## LLM Provider
- **LiteLLM proxy** as the initial provider.
- Requirements: token streaming, temperature/top_p control, model selection.
- Base URL and API key provided via .env.

## Streaming
- Use LiteLLM's streaming (SSE/chunked) with incremental stdout rendering.

## Configuration
- **.env** file for keys and defaults.
- Likely lib: `github.com/joho/godotenv` (tiny), or custom loader.
- Flags override env when provided.
- Suggested keys (draft): `LITELLM_BASE_URL`, `LITELLM_API_KEY`, `LLM_MODEL`.

## Persistence / Memory
- **SQLite** for local history.
- Preferred driver (no CGO): `modernc.org/sqlite`.
- Access via `database/sql` (no heavy ORM).

## Logging
- Prefer stdlib `log/slog` for zero extra deps.

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
- Prefer provider-reported `usage` for tokens if available via LiteLLM.
- Optional local estimation: `github.com/pkoukk/tiktoken-go` (add later if needed).


