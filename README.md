# Tickli

A command line interface for TickTick task management.

## Features

- Smart views: today, tomorrow, next 7 days, inbox, and all tasks across all projects
- Create, update, move, complete, uncomplete, and delete tasks
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

This installs the latest tagged version. If no tags exist, it falls back to the tip of the default branch.

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

# See what's due today (across all projects)
tickli today

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

### Smart Views

| Command              | Description                                              |
| -------------------- | -------------------------------------------------------- |
| `tickli today`       | Show today's tasks and overdue tasks across all projects |
| `tickli tomorrow`    | Show tomorrow's tasks across all projects                |
| `tickli week`        | Show tasks for the next 7 days across all projects       |
| `tickli inbox`       | Show tasks in the inbox                                  |
| `tickli all`         | Show all incomplete tasks across all projects             |

Smart view flags:

| Flag              | Short | Description                                        |
| ----------------- | ----- | -------------------------------------------------- |
| `--all`           | `-a`  | Include completed tasks                            |
| `--priority`      | `-p`  | Filter by minimum priority: `none`, `low`, `medium`, `high` |
| `--tag`           |       | Filter by tag                                      |

### Tasks

| Command                    | Aliases             | Description                         |
| -------------------------- | ------------------- | ----------------------------------- |
| `tickli task create`       | `add`, `a`          | Create a new task                   |
| `tickli task list`         | `ls`                | List tasks                          |
| `tickli task show`         | `info`, `get`       | Show a task                         |
| `tickli task update`       |                     | Update a task                       |
| `tickli task complete`     |                     | Complete a task                     |
| `tickli task uncomplete`   |                     | Uncomplete a task                   |
| `tickli task move`         |                     | Move a task to a different project  |
| `tickli task delete`       | `rm`, `remove`      | Delete a task                       |

Single-task commands (show, update, delete, complete, uncomplete) only need a task ID. The `--project`/`-P` flag is only needed for `list` and `create` (accepts ID or name).

#### Task Flags

| Flag              | Short | Commands         | Description                                        |
| ----------------- | ----- | ---------------- | -------------------------------------------------- |
| `--title`         | `-t`  | create, update   | Task title                                         |
| `--content`       | `-c`  | create, update   | Task content/description                           |
| `--priority`      | `-p`  | create, update, list | Priority: `none`, `low`, `medium`, `high`      |
| `--date`          |       | create, update   | Set date range with natural language               |
| `--start`         |       | create, update   | Start date (ISO 8601)                              |
| `--due`           |       | create, update   | Due date (ISO 8601)                                |
| `--timezone`      |       | create, update   | Timezone for dates                                 |
| `--tags`          |       | create, update   | Comma-separated tags                               |
| `--all-day`       |       | create, update   | Set as all-day task (strips time; `=false` to unset)|
| `--due-within`    |       | list             | Filter: `today`, `tomorrow`, `this-week`, `overdue`|
| `--tag`           |       | list             | Filter by tag                                      |
| `--all`           | `-a`  | list             | Include completed tasks                            |
| `--verbose`       | `-v`  | list             | Show more details                                  |
| `--move-to`/`--to`|       | update           | Move task to a different project (name or ID)      |
| `--to`            |       | move             | Target project (name or ID)                        |
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
| `--project`       | `-P`  | Project context for task list and create (ID or name) |

### Scripting

Tickli is designed to work well in scripts and with other tools:

```bash
# Create a task and capture its ID
TASK_ID=$(tickli task create --title "Review PR" --project inbox --quiet)

# Complete it
tickli task complete "$TASK_ID"

# List task IDs for a project
tickli task list --quiet

# Get task details as JSON
tickli task show <task-id> --json

# Delete a task in one line
tickli task delete $(tickli task create --title "temp" --project inbox --quiet) --force

# Get today's tasks as JSON
tickli today --json

# Count overdue high priority tasks
tickli today -p high --quiet | wc -l

# Create an all-day task (no specific time)
tickli task create -t "Team offsite" --all-day --due "2025-03-20T00:00:00Z"

# Remove all-day status
tickli task update <task-id> --all-day=false

# Date formats accept timezone offsets with or without colon
tickli task create -t "Call" --start "2025-03-14T10:00:00+02:00" --due "2025-03-14T11:00:00+0200"

# Non-interactive fallback: piped/redirected output auto-detects non-TTY
echo "" | tickli today          # prints tab-separated rows instead of fuzzy selector
tickli today | cut -f3          # extract just the title column

# Duration field in JSON: computed from start/due when both are set
tickli task show <task-id> --json | jq .duration
tickli today --json | jq '.[] | select(.duration)'
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
| `default_project`       | string | `"inbox"`   | Default project used when no project is specified (ID or name) |
| `default_project_color` | string | `""`        | Default color for newly created projects         |

You can set the default project with `tickli project use`.

### Files

| Path | Description |
| ---- | ----------- |
| `~/.config/tickli/config.yaml` | Configuration (default project, colors) |
| `~/.config/tickli/credentials` | TickTick Client ID and Client Secret |
| `~/.local/share/tickli/token` | OAuth access token |

To log out, run `tickli reset`. This removes the access token and re-runs authentication.

## Shell Completions

Enable tab completions for commands, flags, project IDs, and task IDs:

```bash
# Bash
tickli completion bash > /etc/bash_completion.d/tickli

# Zsh
tickli completion zsh > "${fpath[1]}/_tickli"

# Fish
tickli completion fish > ~/.config/fish/completions/tickli.fish
```

Restart your shell after installing.

## Design Philosophy

Tickli is built to serve two audiences equally: humans at a terminal and AI agents (or scripts) driving it programmatically.

**Graceful degradation, not separate modes.** When stdin is a TTY, commands present an interactive fuzzy selector. When piped or redirected, they automatically fall back to machine-friendly tab-separated output. No extra flag is needed -the right behavior is inferred from context.

**Consistent, predictable JSON.** Every mutating command (`create`, `update`, `complete`, `uncomplete`, `move`, `delete`) returns the full task object in `--json` mode, not a minimal acknowledgement. Scripts can always parse the same shape. Computed fields like `duration` are included so callers don't need to recompute.

**Flags should be unsurprising.** Each flag name has one meaning across the entire CLI. Where the same concept appears in multiple commands, the flag name and semantics match (e.g. `--to` works on both `move` and `update`). Short flags (`-a`, `-p`, `-t`) are never overloaded within the same command.

**Flexible input, strict output.** Date parsing accepts both `+02:00` and `+0200` timezone offsets because producers vary. Output always uses a single canonical format. `--all-day` automatically strips time components rather than requiring the caller to zero them manually.

**Don't hide the escape hatch.** Boolean flags like `--all-day` support explicit `=false` so that both setting and unsetting are scriptable without needing a separate `--no-all-day` flag.

**Minimal context required.** Single-task commands (`show`, `update`, `delete`, `complete`, `uncomplete`, `move`) only need a task ID. They search all projects automatically. The `--project` flag is only required for `list` and `create`, where a scope is genuinely ambiguous.

**Name or ID, your choice.** Projects can be referenced by ID or name wherever a project argument is accepted (`--project`, `--to`, `--move-to`, positional args). Resolution tries ID first, then falls back to case-insensitive name matching.

**Three-tier output with override hierarchy.** Output comes in three formats: `simple` (human-readable), `json` (machine-readable), and `quiet` (IDs only, for piping). The persistent `--json` and `--quiet` flags override the per-command `-o` flag, so a wrapper script can force JSON globally without modifying individual commands.

**Semantic exit codes.** Exit codes distinguish between success (0), general errors (1), usage errors (2), not-found (3), and auth failures (4). Scripts can branch on the exit code without parsing stderr.

## Documentation

For complete documentation on all available commands:

```bash
tickli --help
```

