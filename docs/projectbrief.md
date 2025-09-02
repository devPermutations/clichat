## Project Brief

**One-liner**: A simple, streaming CLI LLM chat bot written in Go, configured via .env, with lightweight local memory (SQLite or similar), built for rapid iteration.

## Goals
- **Streaming replies**: Tokens render live in the terminal.
- **Minimal setup**: .env-driven config; sane defaults; no daemon.
- **Lightweight memory**: Local persistence for chat history and context.
- **Extensible**: Swappable LLM providers and memory backends.
- **Tooling-ready**: Provider-native tool pass-through (e.g., `web_search`); no local tools in MVP.

## Non-Goals (MVP)
- Rich TUI, web UI, or agents/tools beyond a basic chat loop.
- Multi-user server, auth, or cloud backend.
- Heavy ORMs or complex infra.

## Users
- Developers who want a fast, local-first CLI assistant with streaming.

## MVP Scope
- Single binary CLI.
- LiteLLM proxy adapter as the first provider.
- Local conversation history.
- Simple commands: start chat, view history, clear.
- Slash commands: `/models`, `/model <name>` with tab completion.
- System prompt override via flag/env and in-session command.

## Constraints
- Primary platforms: Docker/Ubuntu and Windows.
- Keep dependencies small; prefer stdlib.
- Avoid CGO if feasible.
- Config and data live in the project folder for containerization.

## Success Criteria
- Start a conversation and receive streamed tokens.
- Persist and retrieve conversation history locally.
- Configure via .env without code changes.
- Streamed output appears naturally formatted (webchat-like, not raw debug tokens).
- Show percentage of context window used; no hard cap enforced.
- Runs in a Docker container on Ubuntu.

## Open Questions (to be resolved via Q&A)
- LiteLLM model selection and listing strategy (e.g., from /models).
- Which CLI framework (stdlib flags vs cobra vs urfave/cli)?
- Exact memory backend (modernc SQLite vs mattn SQLite vs file)?
- Default system prompt and prompt customization?
- Conversation scoping (default conversation id, limits, retention)?
- Env naming conventions and flag overrides?
- Enabling provider-native browsing via LiteLLM (env/flags behavior).


