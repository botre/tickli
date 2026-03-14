# Tickli

A beautiful command line interface for TickTick task management, built with [Charm](https://charm.sh).

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

# Switch to a project
tickli project use

# Create a task
tickli task create -t "Finish documentation"

# Create a task interactively
tickli task create -i

# See what's due today
tickli today

# Complete a task
tickli task complete <task-id>
```

## Commands

### General

| Command           | Description                              |
| ----------------- | ---------------------------------------- |
| `tickli init`     | Initialize tickli                        |
| `tickli reset`    | Reset authentication                     |
| `tickli version`  | Show the version                         |

### Smart Views

| Command              | Description                                              |
| -------------------- | -------------------------------------------------------- |
| `tickli today`       | Show today's tasks and overdue tasks across all projects |
| `tickli tomorrow`    | Show tomorrow's tasks across all projects                |
| `tickli week`        | Show tasks for the next 7 days across all projects       |
| `tickli inbox`       | Show tasks in the inbox                                  |
| `tickli all`         | Show all incomplete tasks across all projects             |

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

Single-task commands (show, update, delete, complete, uncomplete, move) only need a task ID — they search all projects automatically. The `--project`/`-P` flag is only needed for `list` and `create`.

| Flag              | Short | Commands         | Description                                        |
| ----------------- | ----- | ---------------- | -------------------------------------------------- |
| `--title`         | `-t`  | create, update   | Task title                                         |
| `--content`       | `-c`  | create, update   | Task content/description                           |
| `--priority`      | `-p`  | create, update, list | Priority: `none`, `low`, `medium`, `high`      |
| `--date`          |       | create, update   | Set date range with natural language               |
| `--start`         |       | create, update   | Start date (natural language, plain date, or ISO 8601)  |
| `--due`           |       | create, update   | Due date (natural language, plain date, or ISO 8601)    |
| `--timezone`      |       | create, update   | Timezone for dates                                 |
| `--tag`           |       | create, update, list | Comma-separated tags (create/update), filter by tag (list) |
| `--all-day`       |       | create, update   | Set as all-day task (strips time; `=false` to unset)|
| `--due-within`    |       | list             | Filter: `today`, `tomorrow`, `this-week`, `overdue`|
| `--all`           | `-a`  | list             | Include completed tasks                            |
| `--to`            |       | move, update     | Target project for move (required on move)         |
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
| `tickli project delete`  |               | Delete a project                |

| Flag              | Short | Commands         | Description                                        |
| ----------------- | ----- | ---------------- | -------------------------------------------------- |
| `--name`          | `-n`  | create, update   | Project name                                       |
| `--color`         | `-C`  | create, update   | Project color (hex format, e.g. `#F18181`)         |
| `--view-mode`     |       | create, update   | Display mode: `list`, `kanban`, or `timeline`      |
| `--kind`          |       | create, update   | Project type: `TASK` or `NOTE`                     |
| `--force`         | `-f`  | delete           | Skip confirmation prompt                           |
| `--interactive`   | `-i`  | create, update   | Use interactive prompts                            |
| `--filter`        |       | list             | Filter projects by name                            |

### Global Flags

| Flag              | Short | Description                            |
| ----------------- | ----- | -------------------------------------- |
| `--json`          |       | Output in JSON format (overrides `-o`) |
| `--quiet`         | `-q`  | Only print IDs (overrides `-o`)        |
| `--output`        | `-o`  | Output format: `simple` or `json`      |
| `--no-color`      |       | Disable color output (also respects `NO_COLOR` env) |
| `--project`       | `-P`  | Project context for task list and create |

## Interactive Mode

List commands (`today`, `task list`, `project list`, `project use`) open a table picker when running in a terminal. Press `/` to filter, `↑↓` to navigate, `Enter` to select, `Esc` to cancel.

Commands that support `--interactive` (`-i`) use multi-step form wizards:

```bash
tickli task create -i      # project → title → content → priority → date → tags
tickli task update <id> -i
tickli project create -i   # name → color/view mode → type
```

The task creation wizard includes a project selector with type-to-filter and inline "+ Create new project", a tag multi-select from existing tags across all projects, and an inline input for new tags.

Destructive operations (delete, reset) use styled confirmation prompts.

## Scripting

When piped or redirected, list commands automatically output tab-separated columns instead of opening a picker:

| Command         | Columns                                       |
| --------------- | --------------------------------------------- |
| Smart views     | `id`, `[project]`, `title`, `priority`, `due` |
| `task list`     | `id`, `title`, `priority`, `due`              |
| `project list`  | `id`, `name`                                  |

```bash
# Create a task and capture its ID
TASK_ID=$(tickli task create -t "Review PR" -q)

# Complete it
tickli task complete "$TASK_ID"

# Get today's tasks as JSON
tickli today --json

# Count overdue high priority tasks
tickli today -p high -q | wc -l

# Extract just titles
tickli today | cut -f3

# Duration field in JSON
tickli task show <task-id> --json | jq .duration
```

## Configuration

Configuration is stored at `~/.config/tickli/config.yaml` ([XDG](https://specifications.freedesktop.org/basedir-spec/latest/)).

| Key                     | Type   | Default     | Description                                      |
| ----------------------- | ------ | ----------- | ------------------------------------------------ |
| `default_project`       | string | `"inbox"`   | Default project when none specified               |
| `default_project_color` | string | `""`        | Default color for new projects                   |

Set the default project with `tickli project use`.

| Path | Description |
| ---- | ----------- |
| `~/.config/tickli/config.yaml` | Configuration |
| `~/.config/tickli/credentials` | Client ID and Secret |
| `~/.local/share/tickli/token` | OAuth access token |

## Shell Completions

```bash
# Bash
tickli completion bash > /etc/bash_completion.d/tickli

# Zsh
tickli completion zsh > "${fpath[1]}/_tickli"

# Fish
tickli completion fish > ~/.config/fish/completions/tickli.fish
```

## Exit Codes

| Code | Meaning         |
| ---- | --------------- |
| 0    | Success         |
| 1    | General error   |
| 2    | Usage error     |
| 3    | Not found       |
| 4    | Auth error      |

## Design Philosophy

Tickli serves two audiences equally: humans at a terminal and scripts/agents driving it programmatically.

**Graceful degradation.** TTY gets interactive pickers; piped output gets tab-separated columns. No flag needed — inferred from context.

**Consistent JSON.** Every mutating command returns the full object in `--json` mode. Computed fields like `duration` are included.

**Unsurprising flags.** Each flag name has one meaning across the entire CLI. Short flags are never overloaded within a command.

**Flexible input, strict output.** Dates accept plain dates, ISO 8601, natural language. Output uses a single canonical format.

**Minimal context.** Single-entity commands only need an ID. `--project` is only required where scope is ambiguous (`list`, `create`). Smart views need no arguments.

**Three-tier output.** `simple` (human), `json` (machine), `quiet` (IDs). `--json` and `--quiet` override per-command `-o`.

**Stdout is data, stderr is diagnostics.** Piped data stays clean.

**Non-blocking.** Without a TTY, commands never block for input. Every promptable value has a flag.

**One theme, everywhere.** A single `Theme` struct styles everything — pickers, forms, detail views, status messages.
