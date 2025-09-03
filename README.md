Clichat

Quick start
- go build ./cmd/clichat
- Create a .env with your LiteLLM settings
- Run: ./clichat chat

In-session commands (within `chat`)
- /models — list models
- /model <name> — set default model (persists in state.json)
- /history — print recent messages
- /clear — clear messages and reset context stats
- /contextwindow — show prompt/answer counts and token usage

Model management
- List models: ./clichat models
- Set default: ./clichat model <name>

Docker
- Build: docker build -t clichat .
- Run (mount .env and data):
  docker run --rm -it -v "$PWD:/app" --workdir /app clichat chat