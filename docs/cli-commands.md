# CLI Commands Reference

## Overview

Budget-Assist CLI provides a comprehensive set of commands for managing your personal finances. This document covers all available commands and their usage.

## Global Flags

Available for all commands:

```bash
--config string     Config file path (default "~/.budgetassist/config.yaml")
--debug            Enable debug logging
--profile string   Configuration profile to use
--quiet           Suppress all output except errors
```

## Command Categories

### 1. Basic Commands

#### version
Shows the application version and build information.
```bash
budget-assist version
```

#### help
Displays help about any command.
```bash
budget-assist help [command]
```

### 2. Configuration Commands

#### config list
Lists all configuration settings.
```bash
budget-assist config list
```

#### config get
Gets the value of a specific setting.
```bash
budget-assist config get [setting-path]

# Example
budget-assist config get database.path
```

#### config set
Sets the value of a configuration setting.
```bash
budget-assist config set [setting-path] [value]

# Example
budget-assist config set logging.level debug
```

#### config reset
Resets a setting to its default value.
```bash
budget-assist config reset [setting-path]
```

### 3. Import Commands

#### import file
Imports transactions from a file.
```bash
budget-assist import file [file-path] [flags]

Flags:
  --format string    File format (csv|pdf|qr) (default "auto")
  --date-format string    Date format in the file (default "2006-01-02")
  --currency string      Currency code (default "USD")
```

#### import scan
Scans and imports a physical document using the camera.
```bash
budget-assist import scan [flags]

Flags:
  --device int      Camera device number (default 0)
  --type string     Document type (receipt|invoice|statement)
```

### 4. Transaction Management

#### transactions list
Lists transactions with optional filtering.
```bash
budget-assist transactions list [flags]

Flags:
  --from string         Start date (YYYY-MM-DD)
  --to string          End date (YYYY-MM-DD)
  --category string    Filter by category
  --min float         Minimum amount
  --max float         Maximum amount
  --format string     Output format (table|json|csv) (default "table")
```

#### transactions add
Adds a new transaction manually.
```bash
budget-assist transactions add [flags]

Flags:
  --date string       Transaction date (YYYY-MM-DD)
  --amount float     Transaction amount
  --category string  Transaction category
  --description string   Transaction description
  --tags strings        Transaction tags
```

#### transactions edit
Edits an existing transaction.
```bash
budget-assist transactions edit [transaction-id] [flags]

Flags:
  --date string       New date (YYYY-MM-DD)
  --amount float     New amount
  --category string  New category
  --description string   New description
```

### 5. Category Management

#### categories list
Lists all categories.
```bash
budget-assist categories list [flags]

Flags:
  --format string   Output format (table|json) (default "table")
```

#### categories add
Adds a new category.
```bash
budget-assist categories add [name] [flags]

Flags:
  --parent string    Parent category name
  --budget float    Monthly budget amount
  --color string    Category color (hex)
```

### 6. Report Commands

#### report generate
Generates financial reports.
```bash
budget-assist report generate [report-type] [flags]

Flags:
  --from string      Start date (YYYY-MM-DD)
  --to string       End date (YYYY-MM-DD)
  --format string   Output format (pdf|html|json) (default "pdf")
  --output string   Output file path
```

#### report export
Exports data in various formats.
```bash
budget-assist report export [flags]

Flags:
  --format string    Export format (csv|json|xlsx) (default "csv")
  --output string    Output file path
  --what string      Data to export (transactions|categories|all)
```

### 7. Database Management

#### db migrate
Runs database migrations.
```bash
budget-assist db migrate [up|down] [flags]

Flags:
  --steps int    Number of migrations to apply (default all)
```

#### db backup
Creates a database backup.
```bash
budget-assist db backup [flags]

Flags:
  --output string   Backup file path
  --compress       Compress the backup
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid usage |
| 3 | Configuration error |
| 4 | Database error |
| 5 | Import error |

## Examples

### Basic Usage

```bash
# Import a bank statement
budget-assist import file statement.pdf --format pdf

# List recent transactions
budget-assist transactions list --from 2024-01-01

# Generate monthly report
budget-assist report generate monthly --from 2024-01-01 --to 2024-01-31
```

### Advanced Usage

```bash
# Import and categorize automatically
budget-assist import file statement.pdf --auto-categorize

# Export filtered transactions
budget-assist report export --what transactions --format csv \
  --from 2024-01-01 --to 2024-01-31 \
  --output transactions.csv

# Backup database with compression
budget-assist db backup --output backup.sql.gz --compress
```

## Environment Variables

Each flag can be set using environment variables with the prefix `BUDGET_ASSIST_`:

```bash
export BUDGET_ASSIST_CONFIG=/custom/config.yaml
export BUDGET_ASSIST_DEBUG=true
```

## Shell Completion

Generate shell completion scripts:

```bash
# Bash
budget-assist completion bash > /etc/bash_completion.d/budget-assist

# Zsh
budget-assist completion zsh > "${fpath[1]}/_budget-assist"

# Fish
budget-assist completion fish > ~/.config/fish/completions/budget-assist.fish
``` 