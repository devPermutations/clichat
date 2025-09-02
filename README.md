Clichat

Quick start
- go build ./cmd/clichat
- Create a .env with your LiteLLM settings
- Run: ./clichat chat

Bash completion
- Generate: ./clichat completion bash > ~/.config/clichat/completion.bash
- Source it: source ~/.config/clichat/completion.bash

Model management
- List models: ./clichat models
- Set default: ./clichat model <name>

Docker
- Build: docker build -t clichat .
- Run (mount .env and data):
  docker run --rm -it -v "$PWD:/app" --workdir /app clichat chat