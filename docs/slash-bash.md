## Slash Command: `/bash`

### Overview
- **Goal**: Allow users to execute local Bash commands from within the interactive `chat` session using a slash command (`/bash`).
- **Scope**: Non-interactive, short-running commands with stdout/stderr streamed to the terminal.
- **Gate**: Disabled by default; enabled only when `ALLOW_LOCAL_SHELL=true` in `.env`.

### User Experience
- **Syntax**: `/bash <command>`
  - Examples:
    - `/bash echo hello`
    - `/bash ls -la`
    - `/bash docker ps`
- **Autocomplete**:
  - Add `/bash ` to the in-session completer suggestions when `ALLOW_LOCAL_SHELL=true`.
- **Output**:
  - Stream command stdout and stderr directly to the terminal as the process runs.
  - On success, return to prompt normally.
  - On error, print `bash error: <err>` and return to prompt.
- **Cancelation**:
  - Ctrl+C should abort the running command (SIGINT) and return control to the chat loop.

### Configuration & Environment
- `.env` key: `ALLOW_LOCAL_SHELL` (default: `false`)
- Optional (future): `BASH_COMMAND_TIMEOUT_SECS` for command timeout (default: 60s). If not added now, hardcode 60s.

### Execution Strategy
- **OS handling**:
  - Windows: execute via `wsl bash -lc "<command>"`.
  - Linux/macOS: execute via `bash -lc "<command>"`.
- **Working directory**: current process working directory (project root when running `clichat`).
- **Timeout**: Run with `context.WithTimeout` (60s default). Print a friendly timeout message if exceeded.
- **Streaming**: Wire `cmd.Stdout` and `cmd.Stderr` to `os.Stdout` and `os.Stderr`.
- **Return code**: Non-zero exit code should be surfaced as an error message and not crash the CLI.

### Security Considerations
- Disabled by default; explicit opt-in via `ALLOW_LOCAL_SHELL=true`.
- Execute only user-typed commands; do not construct from model output.
- Do not pass secrets or env values to logs; avoid echoing env in error paths.
- Consider a safety notice on first invocation per session when enabled.

### Implementation Plan
1) Config
   - Update `internal/config/config.go` to include:
     - `AllowLocalShell bool` mapped to `ALLOW_LOCAL_SHELL` (default false).
2) CLI – completion
   - In `internal/cli/chat.go` completer, include `/bash ` in `allCmds` only when `cfg.AllowLocalShell` is true.
3) CLI – command handler
   - In `handleSlashCommand(...)` add a new case for `/bash`:
     - Validate `ALLOW_LOCAL_SHELL` is true; otherwise print: `local shell is disabled (set ALLOW_LOCAL_SHELL=true)`.
     - Validate syntax: if no command provided, print `usage: /bash <command>`.
     - Build `exec.CommandContext` with timeout.
       - If `runtime.GOOS == "windows"`: `exec.CommandContext(ctx, "wsl", "bash", "-lc", cmdStr)`
       - Else: `exec.CommandContext(ctx, "bash", "-lc", cmdStr)`
     - Set `Stdout`/`Stderr` to process stdio and run.
     - On error: `fmt.Println("bash error:", err)`.
     - Return `(true, nil)` to indicate handled.

### Example Code Snippet (handler core)
```go
case "/bash":
    if !cfg.AllowLocalShell {
        fmt.Println("local shell is disabled (set ALLOW_LOCAL_SHELL=true)")
        return true, nil
    }
    if len(parts) < 2 {
        fmt.Println("usage: /bash <command>")
        return true, nil
    }
    cmdStr := strings.TrimSpace(strings.TrimPrefix(line, "/bash "))
    if cmdStr == "" {
        fmt.Println("usage: /bash <command>")
        return true, nil
    }
    execCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
    defer cancel()
    var c *exec.Cmd
    if runtime.GOOS == "windows" {
        c = exec.CommandContext(execCtx, "wsl", "bash", "-lc", cmdStr)
    } else {
        c = exec.CommandContext(execCtx, "bash", "-lc", cmdStr)
    }
    c.Stdout = os.Stdout
    c.Stderr = os.Stderr
    if err := c.Run(); err != nil {
        if execCtx.Err() == context.DeadlineExceeded {
            fmt.Println("bash error: command timed out")
        } else {
            fmt.Println("bash error:", err)
        }
    }
    return true, nil
```

### Testing
- Unit test for command construction logic (Windows vs others) using GOOS override or build tags.
- Integration test (opt-in) executing a safe command (e.g., `echo`) behind an env guard, skipped in CI by default.
- Manual QA on Windows (WSL) and Linux/macOS.

### Acceptance Criteria
- With `ALLOW_LOCAL_SHELL=false`, `/bash` warns and does nothing.
- With `ALLOW_LOCAL_SHELL=true`, `/bash echo hi` prints `hi` and returns to prompt.
- Long commands are terminated after the timeout with a clear message.
- Errors do not crash the CLI; control returns to the chat loop.

### Follow-ups (optional)
- Add `BASH_COMMAND_TIMEOUT_SECS` env to configure timeout.
- Persist an opt-in acknowledgment in state to avoid repeating the safety notice.

