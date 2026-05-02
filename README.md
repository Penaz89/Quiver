```text
 ╔═╗ ╦ ╦ ╦ ╦  ╦ ╔═╗ ╦═╗
 ║ ║ ║ ║ ║ ╚╗╔╝ ╠═  ╠╦╝
 ╚═╣ ╚═╝ ╩  ╚╝  ╚═╝ ╩╚═
```

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![Docker](https://img.shields.io/badge/Docker-Alpine-2496ED?logo=docker)](https://hub.docker.com/_/alpine)

An SSH-accessible TUI application built with [Wish](https://github.com/charmbracelet/wish) and [Bubble Tea](https://github.com/charmbracelet/bubbletea), running in a stateless Alpine Linux container.

## Architecture

```
┌───────────────────────────────────────────┐
│          Alpine Linux Container           │
│                                           │
│  ┌────────────────────────────────────┐   │
│  │         Quiver (Go binary)         │   │
│  │                                    │   │
│  │  Wish SSH Server (:2222)           │   │
│  │    └─ Bubble Tea TUI middleware    │   │
│  └────────────────────────────────────┘   │
│                   │                       │
│                   ▼                       │
│         ┌──────────────────┐              │
│         │  /data (volume)  │◄── Persistent│
│         │  • host keys     │    storage   │
│         │  • app data      │              │
│         └──────────────────┘              │
└───────────────────────────────────────────┘
```

The container is **fully stateless** — all persistent data (SSH host keys, application data) is stored on a mounted Docker volume at `/data`. You can destroy and recreate the container at any time without data loss.

## Quick Start

### With Docker Compose (recommended)

```bash
# Build and start
docker compose up -d

# Connect via SSH
ssh localhost -p 2222

# View logs
docker compose logs -f

# Stop
docker compose down
```

### With Docker CLI

```bash
# Build
docker build -t quiver --build-arg VERSION=0.1.0 .

# Run with a named volume
docker run -d \
  --name quiver \
  -p 2222:2222 \
  -v quiver_data:/data \
  quiver

# Connect
ssh localhost -p 2222
```

### Local Development (no container)

```bash
# Install dependencies
go mod download

# Run locally
QUIVER_DATA_DIR=./data go run .

# Connect
ssh localhost -p 2222
```

## Configuration

All configuration is done via environment variables:

| Variable | Default | Description |
|---|---|---|
| `QUIVER_HOST` | `0.0.0.0` | Listen address |
| `QUIVER_PORT` | `2222` | SSH listen port |
| `QUIVER_DATA_DIR` | `/data` | Persistent data directory |

## Project Structure

```
Quiver/
├── main.go              # Application entry point & SSH server setup
├── tui/
│   └── model.go         # Bubble Tea TUI model & views
├── Dockerfile           # Multi-stage Alpine build
├── docker-compose.yml   # Container orchestration with volume
├── .dockerignore
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE              # GPL-3.0
└── README.md
```

## Volume & Persistence

The `/data` volume contains:
- **`host_key_ed25519`** — SSH host key (auto-generated on first run, persisted so clients don't get host key warnings after container recreation)
- Application data (future)

To backup:
```bash
docker run --rm -v quiver_data:/data -v $(pwd):/backup alpine tar czf /backup/quiver-backup.tar.gz -C /data .
```

To restore:
```bash
docker run --rm -v quiver_data:/data -v $(pwd):/backup alpine tar xzf /backup/quiver-backup.tar.gz -C /data
```

## SSH Client Configuration

To avoid `known_hosts` issues during development, add this to `~/.ssh/config`:

```
Host localhost
    UserKnownHostsFile /dev/null
    StrictHostKeyChecking no
```

## License

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

See [LICENSE](LICENSE) for the full text.
