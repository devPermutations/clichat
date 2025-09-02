## Implementation Plan

## Milestones
1. Repo init and docs scaffold.
2. Go module init; CLI skeleton; basic `chat` command.
3. Config loader (.env + flags; precedence: flags > env > defaults).
4. Provider interface + LiteLLM adapter (streaming, models listing).
5. Stream renderer for stdout with minimal buffering and natural formatting.
6. Memory store: SQLite schema + CRUD via `database/sql`.
7. Context budgeter: display % of context window used (provider usage first).
8. Slash commands: `/models`, `/model <name>` with in-session tab completion.
9. Provider-native tools only: pass `tools` to LiteLLM when enabled; no local tool execution.
10. Tests: provider mocks; chat service; model listing.
11. Docker: multi-stage build for Linux container.
12. Build and basic packaging; usage examples.

## Commands (draft)
- `clichat` (interactive chat; default conversation)
- `clichat --conversation mytopic` (named thread)
- `clichat history [--conversation id] [--limit N]`
- `clichat clear [--conversation id]`
- `clichat config test` (validate env/keys)
- Slash in-session: `/model <name>` to switch and persist default model
- Slash in-session: `/models` to list available models

## .env (example draft)
```
# Provider
LLM_PROVIDER=litellm
LITELLM_API_KEY=
LITELLM_BASE_URL=http://localhost:4000
# Model (optional; if empty we select the first from /models)
LLM_MODEL=

# Generation
TEMPERATURE=0.2
TOP_P=1.0

# Storage
DB_PATH=clichat.db

# Prompt
SYSTEM_PROMPT=You are a concise, helpful CLI assistant.

# Provider-native tools
ENABLE_PROVIDER_WEBSEARCH=false

# Context window (optional override)
MODEL_CONTEXT_TOKENS=
```

## Model Defaulting
- If `LLM_MODEL` is unset, on startup fetch `/models` from LiteLLM and use the first available model.
- `/model <name>` updates the default in memory and persists to local config/state.

## Completion
- Generate shell completion scripts via `cobra` (bash first; zsh/fish/powershell later).
- In-session tab completion for `/model <name>` using live results from `/models`.

## Docker
- Multi-stage build producing a static binary (CGO disabled) for Linux.
- Container runs with config and DB in the working directory.

## Tools Strategy
- MVP supports provider-native browsing only (e.g., `web_search`) via LiteLLM; enable on provider/LiteLLM, and the CLI passes through the `tools` parameter when `ENABLE_PROVIDER_WEBSEARCH=true`.

## Testing Notes
- Mock provider stream for determinism.
- Avoid real network in CI.

## Risk/Trade-offs
- Choosing SQLite driver: prefer `modernc.org/sqlite` to avoid CGO on Windows.
- Keep CLI deps minimal; may choose stdlib `flag` first.


