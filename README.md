# Tickli

A command line interface for TickTick task management.

## Features

- Create and manage tasks from your terminal
- Organize tasks into projects
- Set dates, priorities, and tags
- Complete and uncomplete tasks
- Filter tasks by priority, tags, and due date
- Machine-readable output (JSON, quiet mode) for scripting and automation
- Secure OAuth authentication

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
# Initialize and authenticate with TickTick
tickli init

# List available projects
tickli project list

# Switch to a project
tickli project use

# Add a new task
tickli task add "Finish documentation"

# Add a high priority task
tickli task add "Important meeting" -p high --date "tomorrow at 2pm"

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
| `tickli reset`    | Reset authentication (`-f` to skip confirmation) |
| `tickli version`  | Show the current version                 |

### Tasks

| Command                    | Aliases             | Description                         |
| -------------------------- | ------------------- | ----------------------------------- |
| `tickli task create`       | `add`, `a`          | Create a new task                   |
| `tickli task list`         | `ls`                | List and filter tasks               |
| `tickli task show`         | `info`, `get`       | View task details                   |
| `tickli task update`       |                     | Modify a task                       |
| `tickli task complete`     |                     | Mark a task as done                 |
| `tickli task uncomplete`   |                     | Mark a completed task as active     |
| `tickli task delete`       | `rm`, `remove`      | Delete a task                       |

Single-task commands (show, update, delete, complete, uncomplete) only need a task ID. The `-P` flag is only needed for `list` and `create`.

Common task flags:
- `-t` title, `-c` content
- `-p` priority (`none`, `low`, `medium`, `high`)
- `--date`, `--start-date` for dates, `--timezone` for timezone
- `--tags` for tags
- `--due` filter on list (`today`, `tomorrow`, `this-week`, `overdue`)

### Projects

| Command                  | Aliases       | Description                     |
| ------------------------ | ------------- | ------------------------------- |
| `tickli project list`    | `ls`          | List and browse projects        |
| `tickli project show`    | `info`, `get` | View project details            |
| `tickli project create`  |               | Create a new project            |
| `tickli project use`     |               | Switch active project context   |
| `tickli project update`  |               | Modify a project                |
| `tickli project delete`  | `rm`          | Delete a project                |

### Global Flags

| Flag              | Short | Description                            |
| ----------------- | ----- | -------------------------------------- |
| `--json`          |       | Output in JSON format (shorthand for `-o json`) |
| `--quiet`         | `-q`  | Only print IDs (useful for piping)     |
| `--no-color`      |       | Disable color output                   |
| `-o`, `--output`  |       | Output format: `simple`, `json`        |

### Scripting

Tickli is designed to work well in scripts and with other tools:

```bash
# Create a task and capture its ID
TASK_ID=$(tickli task create -t "Review PR" -P inbox -q)

# Complete it
tickli task complete "$TASK_ID"

# List task IDs for a project
tickli task list -q

# Get task details as JSON
tickli task show <task-id> --json
```

### Exit Codes

| Code | Meaning         |
| ---- | --------------- |
| 0    | Success         |
| 1    | General error   |
| 2    | Usage error     |
| 3    | Not found       |
| 4    | Auth error      |

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

To log out, run `tickli reset`. This removes the access token and re-runs authentication.

## Documentation

For complete documentation on all available commands:

```bash
tickli --help
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.