```text
 ╔═╗ ╦ ╦ ╦ ╦  ╦ ╔═╗ ╦═╗
 ║ ║ ║ ║ ║ ╚╗╔╝ ╠═  ╠╦╝
 ╚═╣ ╚═╝ ╩  ╚╝  ╚═╝ ╩╚═
```

<div align="center">
  <p><b>An advanced SSH-accessible Terminal User Interface (TUI) application for personal and family management.</b></p>
  <p>
    <a href="https://www.gnu.org/licenses/gpl-3.0"><img src="https://img.shields.io/badge/License-GPLv3-blue.svg?style=for-the-badge" alt="License: GPL v3" /></a>
    <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go" alt="Go" /></a>
    <a href="https://hub.docker.com/_/alpine"><img src="https://img.shields.io/badge/Docker-Alpine-2496ED?style=for-the-badge&logo=docker" alt="Docker" /></a>
  </p>
</div>

---

**Quiver** is a centralized, self-hosted TUI dashboard built with [Wish](https://github.com/charmbracelet/wish), [Bubble Tea](https://github.com/charmbracelet/bubbletea), and [Lipgloss](https://github.com/charmbracelet/lipgloss). Originally created to fulfill a **personal need** for a robust local server application, it is specifically designed to be **deployed via Docker Compose** and accessed remotely using **[NeXterm](https://nexterm.dev/)** (or any standard SSH client) for a secure, distraction-free management experience on any home server.

---

## ✨ Features

Quiver packs a wide array of tools to manage your daily life, finances, and tasks straight from the terminal.

### 💰 Comprehensive Financial Tracking
- **Salary Tracking**: Keep track of Gross, Net, Deductions, and Tax percentages. Features year-over-year bar chart comparisons.
- **Expense Management**: Categorize and track *Recurring* and *Daily* expenses, complete with monthly bar chart analytics and impact calculation against your net salary.
- **Transfers & Installments**: Manage internal money transfers (Giroconti) and track multi-month installments with ease.

### 📋 Productivity & Organization
- **Dynamic Dashboard**: A fully responsive landing page summarizing your financial totals, upcoming deadlines, recent journal entries, and tasks.
- **GTD Task Board**: A Kanban-style "Getting Things Done" workflow (`TODO`, `DOING`, `DONE`) featuring priority markers, projects, and deadlines.
- **Habit Tracker**: Build good habits with a "Don't Break The Chain" visual ASCII heatmap.
- **Daily Journal**: A secure, date-based plain-text journal with automated Markdown export capabilities.

### 👥 Collaboration & Multi-User
- **Family Workspaces**: Create shared workspaces to collaborate with family members. Switch between personal and family contexts instantly (`Ctrl+E`). Shared entries automatically track the creator's username.
- **Real-Time Chat**: An integrated real-time chat interface for all users within the same Quiver instance.
- **Admin Control Panel**: Manage users (create, edit, securely delete) via an exclusive administrative interface.

### 🚗 Lifestyle & Utilities
- **Vehicle Management**: Track services, road tax, and insurance deadlines, automatically synced to the Home Dashboard.
- **Live Weather**: Integrated weather widget utilizing `wttr.in`.

### 🎨 Theming & Localization
- **UI & Theming**: Dynamically switchable color palettes (e.g., *Catppuccin*, *Nord*, *Gruvbox*). The UI features smart focus dimming and responsive layouts optimized for varying terminal sizes.
- **i18n Support**: Full support for both **English** and **Italian** languages.

---

## 🏗️ Architecture

The application is completely **stateless and portable**. All data is serialized as JSON files and stored safely on a mounted volume.

```text
┌───────────────────────────────────────────┐
│          Alpine Linux Container           │
│                                           │
│  ┌────────────────────────────────────┐   │
│  │         Quiver (Go binary)         │   │
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

---

## 🚀 Deployment

Quiver is cross-platform and host-agnostic, perfect for a NAS, Raspberry Pi, or any standard Linux server.

### 1. Portainer / Docker Compose (Recommended)

Quiver is optimized for Portainer Stack deployments.

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

1. In Portainer, navigate to **Stacks** -> **Add stack**.
2. Paste the `docker-compose.yml` above.
3. Click **Deploy the stack**.
4. Connect via terminal: `ssh localhost -p 2222` (Replace `localhost` with your server's IP). 
   *Note: No password is required for the SSH connection. Authentication is handled internally by Quiver.*

### 2. Docker CLI

```bash
# Build the image
docker build -t quiver .

# Run the container
docker run -d \
  --name quiver \
  -p 2222:2222 \
  -v quiver_data:/data \
  --restart unless-stopped \
  quiver
```

### 3. Local Development

```bash
# Install Go dependencies
go mod tidy

# Run the SSH server locally
QUIVER_DATA_DIR=./data go run .

# Connect via SSH in another terminal
ssh localhost -p 2222
```

---

## ⚙️ Configuration

Minimal setup is required. Use environment variables to override defaults:

| Variable | Default | Description |
|---|---|---|
| `QUIVER_HOST` | `0.0.0.0` | IP address the SSH server listens on. |
| `QUIVER_PORT` | `2222` | Port the SSH server listens on. |
| `QUIVER_DATA_DIR` | `/data` | Path to the persistent storage directory. |

---

## 💾 Volume & Persistence

The `/data` volume contains your SSH host keys (to prevent warnings on container recreation) and your data organized as follows:
- `users.json` / `families.json`: Registries for users and family workspaces.
- `users/{username}/`: Personal JSON data stores (finances, tasks, journal, etc.).
- `families/{familyID}/`: Shared JSON data stores for collaboration.
- `themes/`: Drop custom `.json` themes here to load them dynamically.

**Backup & Restore Examples:**
```bash
# Backup
docker run --rm -v quiver_data:/data -v $(pwd):/backup alpine tar czf /backup/quiver-backup.tar.gz -C /data .

# Restore
docker run --rm -v quiver_data:/data -v $(pwd):/backup alpine tar xzf /backup/quiver-backup.tar.gz -C /data
```

---

## 🌐 Remote Access

### NeXterm (Recommended)
For the ultimate remote experience, we highly recommend **[NeXterm](https://nexterm.dev/)**. It provides a robust, web-based terminal that handles SSH connections beautifully, preventing shortcut conflicts and ensuring your TUI looks sharp from any browser.

### Standard SSH Tips
During development, to avoid strict host key checking errors when tearing down containers, append this to your `~/.ssh/config`:

```text
Host localhost
    Port 2222
    UserKnownHostsFile /dev/null
    StrictHostKeyChecking no
    LogLevel ERROR
```

---

## 📄 License

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

See [LICENSE](LICENSE) for the full text.

---

<div align="center">
  <p><i>Developed for personal use, shared with the community.</i></p>
</div>
