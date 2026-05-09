```text
 ╔═╗ ╦ ╦ ╦ ╦  ╦ ╔═╗ ╦═╗
 ║ ║ ║ ║ ║ ╚╗╔╝ ╠═  ╠╦╝
 ╚═╣ ╚═╝ ╩  ╚╝  ╚═╝ ╩╚═
```

[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![Docker](https://img.shields.io/badge/Docker-Alpine-2496ED?logo=docker)](https://hub.docker.com/_/alpine)

An advanced SSH-accessible Terminal User Interface (TUI) application built with [Wish](https://github.com/charmbracelet/wish), [Bubble Tea](https://github.com/charmbracelet/bubbletea), and [Lipgloss](https://github.com/charmbracelet/lipgloss). Originally created to fulfill a **personal need** for a centralized self-hosted dashboard, it is specifically designed to be **deployed via Docker Compose** and accessed remotely using **[NeXterm](https://nexterm.dev/)** (or any SSH client) for a secure, terminal-based management experience on any home server.

## Features

- **Multi-user Authentication**: Secure login system with bcrypt encryption. Each user's data is isolated in their own personal directory. Includes account creation and secure logout functionalities.
- **Family Workspaces & Author Tracking**: Create shared workspaces to collaborate with other users on the server. Easily switch between personal and family contexts using `Ctrl+W`. All shared entries (Salaries, Tasks, Daily Expenses) automatically track and display the username of the creator.
- **Admin Control Panel**: A special `admin` user is generated automatically on the first run (the default password is `admin`, valid only for the first login, after which you will be forced to change it). The admin has access to an exclusive menu to list, add, edit, and securely delete users.
- **Dynamic Home Dashboard**: A fully responsive landing page summarizing your financial totals, upcoming deadlines (drawn from Tasks, Vehicles, and Insurances), your latest Journal notes, and your most recently added Tasks.
- **Real-Time Chat**: Integrated chat interface allowing users within the same Quiver instance to communicate in real-time, featuring a clean, responsive layout.
- **GTD Task Management**: A Kanban-style "Getting Things Done" workflow. Track tasks across columns (`TODO`, `DOING`, `DONE`) with priority markers, projects, and deadlines.
- **Habit Tracker**: Track daily habits in a "Don't Break The Chain" style, visualizing progress with a GitHub-style ASCII heatmap.
- **Journal**: A personal plain-text daily journal featuring date-based navigation and automated Markdown export capabilities.
- **Advanced Financial Tracking**: A comprehensive module to track housing, holidays, subscriptions, and **salaries**. Includes dynamic visual comparisons (bar charts for gross/net/taxes) and an automatic "Salary Impact" projection that calculates how fixed expenses weigh on your estimated annual net income.
- **Vehicle Management**: Detailed vehicle tracking including service costs, road tax, and insurance deadlines, automatically synchronized with the Home Dashboard.
- **Live Weather**: Integrated weather widget utilizing `wttr.in`.
- **Theming & UI**: Dynamically switchable color palettes accessible from Settings (includes *Catppuccin*, *Nord*, *Gruvbox*, etc.). The UI features smart focus dimming and fully theme-aware structural borders for an optimal, distraction-free user experience.
- **Localization (i18n)**: Full support for both English and Italian languages, changeable seamlessly from the Settings menu.
- **Stateless & Portable**: Fully Dockerized architecture. Destroy and recreate the container at will; just mount a volume to persist user data, settings, and SSH host keys.

## Architecture

```text
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

The container is **fully stateless**. All application data is serialized in JSON files and stored alongside SSH host keys on a mounted Docker volume at `/data`.

## Deployment

Quiver is built with cross-platform deployment in mind and is completely agnostic to the underlying host, making it perfect for Raspberry Pi or NAS setups. 

### 1. Portainer / Docker Compose (Recommended)

Quiver is heavily optimized to be deployed as a Portainer Stack.

```yaml
services:
  quiver:
    image: quiver:latest
    build: .
    container_name: quiver_app
    ports:
      - "2222:2222"
    volumes:
      - quiver_data:/data
    restart: unless-stopped

volumes:
  quiver_data:
```
**In Portainer:**
1. Navigate to **Stacks** -> **Add stack**.
2. Paste the `docker-compose.yml` above.
3. Click **Deploy the stack**.
4. Connect via terminal: `ssh localhost -p 2222` (replace localhost with your server's IP). **Note:** The SSH connection does not require a password. If prompted, just press enter; authentication is handled internally by the Quiver TUI.

### 2. Docker CLI

If you prefer the command line:

```bash
# Build the image
docker build -t quiver .

# Run the container with a persistent named volume
docker run -d \
  --name quiver \
  -p 2222:2222 \
  -v quiver_data:/data \
  --restart unless-stopped \
  quiver
```

### 3. Local Development (No Docker)

To compile and run the application directly on your host machine:

```bash
# Install Go dependencies
go mod tidy

# Run the SSH server locally
QUIVER_DATA_DIR=./data go run .

# In another terminal window, connect via SSH (no SSH password required)
ssh localhost -p 2222
```

## Configuration

The application requires minimal configuration, handled entirely via environment variables:

| Variable | Default | Description |
|---|---|---|
| `QUIVER_HOST` | `0.0.0.0` | IP address the SSH server listens on. |
| `QUIVER_PORT` | `2222` | Port the SSH server listens on. |
| `QUIVER_DATA_DIR` | `/data` | Path to the persistent storage directory. |

## Volume & Persistence

The mapped `/data` volume will generate the following structure automatically:
- `host_key_ed25519`: The secure SSH host key. Persisting this prevents SSH clients from throwing warnings if you recreate the container.
- `users.json`: The global registry of all registered users and their encrypted passwords.
- `families.json`: The registry of all shared family workspaces and their members.
- `users/{username}/`: Dedicated folders containing all personal JSON-based data stores (habits, journal, tasks, finances, etc.) for that specific user.
- `families/{familyID}/`: Dedicated folders containing all shared JSON-based data stores for family workspaces.
- `themes/`: Directory for placing custom `.json` theme definitions to be loaded dynamically.

### Backing Up Data

```bash
docker run --rm -v quiver_data:/data -v $(pwd):/backup alpine tar czf /backup/quiver-backup.tar.gz -C /data .
```

### Restoring Data

```bash
docker run --rm -v quiver_data:/data -v $(pwd):/backup alpine tar xzf /backup/quiver-backup.tar.gz -C /data
```

## Remote Access & SSH Clients

### NeXterm (Recommended)
For the best experience, especially when accessing Quiver remotely, we recommend using **[NeXterm](https://nexterm.dev/)**. It provides a robust, web-based terminal interface that handles SSH connections beautifully, making it easy to access your dashboard from any browser.

### Standard SSH Client Tips
During testing and development, you might tear down and rebuild the container frequently without persisting the host keys. To prevent `known_hosts` strict checking issues, append this to your `~/.ssh/config`:

```text
Host localhost
    Port 2222
    UserKnownHostsFile /dev/null
    StrictHostKeyChecking no
    LogLevel ERROR
```

## License

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

See [LICENSE](LICENSE) for the full text.
