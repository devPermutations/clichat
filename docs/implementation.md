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


## Step-by-step Execution Guide

1) Bootstrap repo and Go module
- Initialize module and basic scaffolding.
```
mkdir -p cmd/clichat internal/{cli,chat,provider/litellm,config,memory/sqlite,stream,context}
go mod init github.com/yourname/clichat
echo "Clichat" > README.md
```

2) Add dependencies
- Cobra (CLI), Liner (interactive), dotenv, SQLite driver, and testing libs.
```
go get github.com/spf13/cobra@latest
go get github.com/peterh/liner@v1.2.2
go get github.com/joho/godotenv@v1.5.1
go get modernc.org/sqlite@latest
```

3) Layout project structure
```
cmd/
  clichat/
    main.go
internal/
  cli/
    root.go
    chat.go            # interactive loop (liner), slash commands
    completion.go      # shell completion (bash first)
  chat/
    service.go         # orchestrates config, memory, provider, streaming
  provider/
    litellm/
      client.go        # streaming chat, list models, pass-through tools
      types.go
  config/
    config.go          # .env + flags + defaults; validation
  memory/
    sqlite/
      store.go         # schema init and CRUD
  stream/
    renderer.go        # natural streaming to stdout
  context/
    budget.go          # compute % of context used
```

4) Implement config loader
- Precedence: flags > env > defaults. Keys:
  - `LLM_PROVIDER=litellm`, `LITELLM_BASE_URL`, `LITELLM_API_KEY`
  - `LLM_MODEL` (optional), `TEMPERATURE`, `TOP_P`
  - `SYSTEM_PROMPT`, `MODEL_CONTEXT_TOKENS` (optional), `DB_PATH`
  - `ENABLE_PROVIDER_WEBSEARCH` (true|false)

5) Implement LiteLLM client
- Endpoints: `/v1/chat/completions` (stream), `/v1/models` (list).
- Inputs: model, messages, temperature, top_p, tools (only when `ENABLE_PROVIDER_WEBSEARCH=true`).
- Output: streamed deltas; final usage for token accounting when available.

6) Implement memory store (SQLite, modernc)
- On startup, create tables if not exist (`conversations`, `messages`).
- Functions: `CreateOrGetConversation(id)`, `AppendMessage`, `ListMessages(conversationID, limit)`.

7) Implement chat service
- Loads prior messages for selected conversation.
- Applies system prompt (if set), user message, then streams assistant tokens.
- Persists user and assistant messages after completion.

8) Interactive CLI with liner
- `clichat` starts interactive loop (default conversation).
- Slash commands:
  - `/models` → list models from LiteLLM
  - `/model <name>` → set default model (persist locally)
  - `/exit` or Ctrl-D → quit
- Tab completion for `/model <name>` based on live `/models` results.

9) Stream renderer and natural formatting
- Print tokens as they arrive; coalesce whitespace/newlines for natural feel.
- Print footer line with run stats (elapsed, tokens if provided, context %).

10) Context percentage display
- Prefer provider-reported usage to compute %: `(prompt+completion_tokens)/MODEL_CONTEXT_TOKENS`.
- If usage/context is missing, show `N/A` (add local estimation later if desired).

11) Bash completion
- Generate bash completion with cobra and add an installation note in README.
```
clichat completion bash > /etc/bash_completion.d/clichat    # root
# or
clichat completion bash > ~/.config/clichat/completion.bash
source ~/.config/clichat/completion.bash
```

12) Docker packaging
- Multi-stage Dockerfile producing a static Linux binary.
- Mount project folder as working directory; `.env` and DB live alongside binary.

13) Tests
- Provider client: mock HTTP server for `/models` and streaming.
- Chat service: unit tests for message assembly and persistence.
- Memory store: schema and CRUD tests.

14) Acceptance checklist (MVP)
- `clichat` runs, accepts input, streams responses.
- `/models` lists models; `/model` switches and persists default.
- Messages saved and restored; context % shown when available.
- Docker image builds and runs; bash completion works.

15) Nice-to-haves (post-MVP)
- Zsh/fish/PowerShell completion.
- Local token estimation.
- Export/import conversations.
- Additional providers via LiteLLM config.


