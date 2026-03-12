# Tickli

A command line interface for TickTick task management.

## Features

- Create, update, complete, uncomplete, and delete tasks
- Create, update, and delete projects
- Set dates, priorities, tags, and content on tasks
- Filter tasks by priority, tags, and due date
- Interactive fuzzy-search selection for tasks and projects
- Machine-readable output (JSON, quiet mode) for scripting and automation
- Semantic exit codes for reliable error handling
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

### 2. Set Up and Use

```bash
# Initialize and authenticate with TickTick
tickli init

# List available projects
tickli project list

# Switch to a project (interactive selector)
tickli project use

# Switch to a project by ID
tickli project use abc123def456

# Create a task
tickli task create --title "Finish documentation"

# Create a high priority task with a due date
tickli task create --title "Important meeting" --priority high --date "tomorrow at 2pm"

# List your tasks
tickli task list

# Complete a task
tickli task complete <task-id>

# Uncomplete a task
tickli task uncomplete <task-id>
```

## Commands

### General

| Command           | Description                              |
| ----------------- | ---------------------------------------- |
| `tickli init`     | Initialize tickli                        |
| `tickli reset`    | Reset authentication (`--force`/`-f` to skip confirmation) |
| `tickli version`  | Show the version                         |

### Tasks

| Command                    | Aliases             | Description                         |
| -------------------------- | ------------------- | ----------------------------------- |
| `tickli task create`       | `add`, `a`          | Create a new task                   |
| `tickli task list`         | `ls`                | List tasks                          |
| `tickli task show`         | `info`, `get`       | Show a task                         |
| `tickli task update`       |                     | Update a task                       |
| `tickli task complete`     |                     | Complete a task                     |
| `tickli task uncomplete`   |                     | Uncomplete a task                   |
| `tickli task delete`       | `rm`, `remove`      | Delete a task                       |

Single-task commands (show, update, delete, complete, uncomplete) only need a task ID. The `--project-id`/`-P` flag is only needed for `list` and `create`.

#### Task Flags

| Flag              | Short | Commands         | Description                                        |
| ----------------- | ----- | ---------------- | -------------------------------------------------- |
| `--title`         | `-t`  | create, update   | Task title                                         |
| `--content`       | `-c`  | create, update   | Task content/description                           |
| `--priority`      | `-p`  | create, update, list | Priority: `none`, `low`, `medium`, `high`      |
| `--date`          |       | create, update   | Due date                                           |
| `--start-date`    |       | create, update   | Start date                                         |
| `--timezone`      |       | create, update   | Timezone for dates                                 |
| `--tags`          |       | create, update   | Comma-separated tags                               |
| `--all-day`       | `-a`  | create, update   | Set as all-day task                                |
| `--due`           |       | list             | Filter: `today`, `tomorrow`, `this-week`, `overdue`|
| `--tag`           |       | list             | Filter by tag                                      |
| `--all`           | `-a`  | list             | Include completed tasks                            |
| `--verbose`       | `-v`  | list             | Show more details                                  |
| `--force`         | `-f`  | delete           | Skip confirmation prompt                           |
| `--interactive`   | `-i`  | create, update   | Use interactive prompts                            |

### Projects

| Command                  | Aliases       | Description                     |
| ------------------------ | ------------- | ------------------------------- |
| `tickli project list`    | `ls`          | List projects                   |
| `tickli project show`    | `info`, `get` | Show a project                  |
| `tickli project create`  |               | Create a new project            |
| `tickli project use`     |               | Set the active project          |
| `tickli project update`  |               | Update a project                |
| `tickli project delete`  | `rm`          | Delete a project                |

#### Project Flags

| Flag              | Short | Commands         | Description                                        |
| ----------------- | ----- | ---------------- | -------------------------------------------------- |
| `--name`          | `-n`  | create, update   | Project name                                       |
| `--color`         | `-c`  | create, update   | Project color (hex format, e.g. `#F18181`)         |
| `--force`         | `-f`  | delete           | Skip confirmation prompt                           |
| `--interactive`   | `-i`  | create, update   | Use interactive prompts                            |
| `--with-tasks`    |       | show             | Include all tasks                                  |
| `--filter`        |       | list             | Filter projects by name                            |

### Global Flags

| Flag              | Short | Description                            |
| ----------------- | ----- | -------------------------------------- |
| `--json`          |       | Output in JSON format                  |
| `--quiet`         | `-q`  | Only print IDs (useful for piping)     |
| `--output`        | `-o`  | Output format: `simple`, `json`        |
| `--no-color`      |       | Disable color output                   |
| `--project-id`    | `-P`  | Project context for task list and create |

### Scripting

Tickli is designed to work well in scripts and with other tools:

```bash
# Create a task and capture its ID
TASK_ID=$(tickli task create --title "Review PR" --project-id inbox --quiet)

# Complete it
tickli task complete "$TASK_ID"

# List task IDs for a project
tickli task list --quiet

# Get task details as JSON
tickli task show <task-id> --json

# Delete a task in one line
tickli task delete $(tickli task create --title "temp" --project-id inbox --quiet) --force
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

You can set the default project with `tickli project use`.

### Files

| Path | Description |
| ---- | ----------- |
| `~/.config/tickli/config.yaml` | Configuration (default project, colors) |
| `~/.config/tickli/credentials` | TickTick Client ID and Client Secret |
| `~/.local/share/tickli/token` | OAuth access token |

To log out, run `tickli reset`. This removes the access token and re-runs authentication.

## Shell Completions

Tickli supports tab completions for commands, flags, project IDs, and task IDs.

#### Bash

```bash
# Current session
source <(tickli completion bash)

# Persistent (Linux)
tickli completion bash > /etc/bash_completion.d/tickli

# Persistent (macOS)
tickli completion bash > $(brew --prefix)/etc/bash_completion.d/tickli
```

#### Zsh

```bash
# Current session
source <(tickli completion zsh)

# Persistent (Linux)
tickli completion zsh > "${fpath[1]}/_tickli"

# Persistent (macOS)
tickli completion zsh > $(brew --prefix)/share/zsh/site-functions/_tickli
```

> If completions aren't working, ensure `compinit` is loaded in your `~/.zshrc`:
> `autoload -U compinit; compinit`

#### Fish

```bash
# Current session
tickli completion fish | source

# Persistent
tickli completion fish > ~/.config/fish/completions/tickli.fish
```

## Documentation

For complete documentation on all available commands:

```bash
tickli --help
```

