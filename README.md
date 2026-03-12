# Tickli 📋✨

> A modern command line interface for TickTick task management

![Tickli Demo](assets/tickli-demo.gif)

## What is Tickli?

Tickli is a powerful CLI tool that lets you manage your TickTick tasks and projects directly from your terminal. Stay productive without leaving your command line!

## Features

- 🚀 Create and manage tasks right from your terminal
- 📂 Organize tasks into projects
- 📅 Set dates, priorities, and tags
- 🔄 Complete and uncomplete tasks
- 🔍 Filter and search your tasks
- 🔐 Secure OAuth authentication

## Installation

```bash
go install github.com/botre/tickli@latest
```

## Quick Start

### 1. Create a TickTick App

1. Go to the [TickTick Developer Portal](https://developer.ticktick.com/manage)
2. Create a new app
3. Set the **OAuth redirect URL** to `http://localhost:8080`
4. Note your **Client ID** and **Client Secret**

### 2. Authenticate

```bash
# Initialize and authenticate with TickTick (you will be prompted for your Client ID and Secret)
tickli init

# List available projects
tickli project list

# Switch to a project
tickli project use "Work Tasks"

# Add a new task
tickli task add "Finish documentation for project X"

# Add a high priority task due tomorrow
tickli task add "Important meeting" --priority high --date "tomorrow at 2pm"

# List your tasks
tickli task list

# Complete a task
tickli task complete <task-id>
```

## Commands

### General

| Command           | Description                              |
| ----------------- | ---------------------------------------- |
| `tickli init`     | Authenticate with TickTick via OAuth     |
| `tickli reset`    | Reset authentication (use `--force` to skip confirmation) |
| `tickli version`  | Show the current version                 |

### Tasks

| Command                | Aliases             | Description                          |
| ---------------------- | ------------------- | ------------------------------------ |
| `tickli task create`   | `add`, `a`          | Create a new task                    |
| `tickli task list`     | `ls`                | List and browse tasks interactively  |
| `tickli task show`     | `info`, `get`       | View task details                    |
| `tickli task update`   |                     | Modify a task's properties           |
| `tickli task complete` |                     | Mark a task as done                  |
| `tickli task delete`   | `rm`, `remove`      | Delete a task                        |

Common flags: `-t` title, `-c` content, `-p` priority (none/low/medium/high), `--date`/`--start`/`--due` for dates, `--tags`, `--repeat`, `-P` project ID override.

### Projects

| Command                  | Aliases       | Description                     |
| ------------------------ | ------------- | ------------------------------- |
| `tickli project list`    | `ls`          | List and browse projects        |
| `tickli project show`    | `info`, `get` | View project details            |
| `tickli project create`  |               | Create a new project            |
| `tickli project use`     |               | Switch active project context   |
| `tickli project update`  |               | Modify a project's properties   |
| `tickli project delete`  | `rm`          | Delete a project                |

## Configuration

Tickli stores its configuration at `~/.config/tickli/config.yaml` (following the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/latest/)).

| Key                     | Type   | Default     | Description                                      |
| ----------------------- | ------ | ----------- | ------------------------------------------------ |
| `default_project_id`    | string | `""`        | Default project ID used when no project is specified |
| `default_project_color` | string | `"#FF1111"` | Default color for newly created projects         |

You can set the default project interactively with `tickli project use`.

### Files

| Path | Description |
| ---- | ----------- |
| `~/.config/tickli/config.yaml` | Configuration (default project, colors) |
| `~/.config/tickli/credentials` | TickTick Client ID and Client Secret |
| `~/.local/share/tickli/token` | OAuth access token |

To log out, run `tickli reset` — this removes the access token and re-runs authentication.

## Documentation

For complete documentation on all available commands:

```bash
tickli --help
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.