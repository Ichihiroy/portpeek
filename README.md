# portpeek

A tiny terminal UI that shows what's listening on your local TCP ports and lets you kill it.

Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea). Linux and macOS (needs `lsof`).

## Install

```bash
go install github.com/YOUR_USER/portpeek@latest
```

Or build from source:

```bash
git clone https://github.com/YOUR_USER/portpeek
cd portpeek
go build -o portpeek .
```

## Usage

```bash
portpeek
```

| Key       | Action                          |
| --------- | ------------------------------- |
| `↑` / `↓` | Move selection                  |
| `k`       | Kill selected process (SIGTERM) |
| `r`       | Refresh                         |
| `/`       | Filter by port, PID, or name    |
| `esc`     | Clear filter                    |
| `q`       | Quit                            |

## Why

`lsof -i -P -n | grep LISTEN` works, but it's hard to read and you still have to copy a PID and run `kill` by hand. This does both in one place.
