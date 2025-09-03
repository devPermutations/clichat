## Architecture

## High-Level Components
- **CLI Layer**: `cobra` commands, interactive loop with `liner`, slash commands (`/models`, `/model`).
- **Chat Service**: Orchestrates prompt, context, tool definitions, and provider calls.
- **Provider Client**: Interface + LiteLLM adapter (streaming, tools, models listing).
- **Stream Renderer**: Writes tokens to stdout with minimal latency and natural formatting.
- **Memory Store**: Persists conversations/messages locally (SQLite).
- **Context Budgeter**: Tracks tokens used; displays % of model context consumed.
- **Config Loader**: Reads .env via godotenv; no flags yet.
- **Tooling**: Provider-native tool pass-through only; no local tool execution in MVP.
- **Logger**: Minimal stdout/stderr via fmt; slog later.

## Data Flow (MVP)
1. User runs `clichat` and enters a message.
2. Chat Service loads config, prior context (per conversation).
3. Provider Client issues a streamed completion request (with provider-native tools if enabled).
4. Stream Renderer prints tokens as they arrive.
5. If the model uses provider-native browsing via LiteLLM, the provider handles it; the CLI streams output only.
6. Memory Store appends the user and assistant messages to history.
7. Context Budgeter updates and prints % context used.

## Interfaces (sketch)
- ProviderClient: `StreamChat(request) -> (<-chan Token, <-chan error)`
- MemoryStore: `AppendMessage`, `ListMessages(conversationID, limit)`
- ChatService: `HandleUserInput(conversationID, text)`
- Models: `ListModels() ([]Model, error)`

## Data Model (initial)
- `conversations(id TEXT PRIMARY KEY, title TEXT, created_at TIMESTAMP)`
- `messages(id INTEGER PRIMARY KEY AUTOINCREMENT, conversation_id TEXT, role TEXT, content TEXT, created_at TIMESTAMP)`
 
## Configuration Keys (draft)
- `LLM_PROVIDER=litellm`
- `LITELLM_BASE_URL`, `LITELLM_API_KEY`
- `LLM_MODEL` (optional; if empty, default to first available model)
- `TEMPERATURE`, `TOP_P`
- `DB_PATH` (e.g., `clichat.db`)
- `SYSTEM_PROMPT`
- `MODEL_CONTEXT_TOKENS` (optional override for context % computation)
- `ENABLE_PROVIDER_WEBSEARCH=true|false`
 - `DROP_SAMPLING_PARAMS`
 - `DEBUG_PROMPTS`

## Observability
- Minimal structured logs to stderr; redact secrets.

## Security/Privacy
- Do not log API keys or full prompts with PII.
- Store data locally in user space; document location.

## Slash Commands (initial)
- `/model <name>`: Switch and persist default model for subsequent chats.
- `/models`: List models from LiteLLM; supports tab completion in-session.
- `/history`: Print recent messages for the current conversation.
- `/clear`: Clear messages and reset context stats for the current conversation.
- `/contextwindow`: Show prompt/answer counts and token usage; percentage shown when `MODEL_CONTEXT_TOKENS` is set.

## Notes on Websearch
- Provider-native browsing only in MVP: if exposed via LiteLLM (e.g., `web_search`), enable it at the provider/LiteLLM level; the CLI passes through `tools` and does not implement a custom tool.


