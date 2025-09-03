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
- Commands: `chat`, `models`, `model <name>`.
- In-session slash commands: `/models`, `/model <name>`, `/history`, `/clear`, `/contextwindow`.
- System prompt override via env (`SYSTEM_PROMPT`).

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
- Default conversation scoping beyond `default` (naming, limits, retention).
- Token accounting refinements and potential provider-reported usage integration.


