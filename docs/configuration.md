# Configuration Guide

## Overview

Budget-Assist uses a hierarchical configuration system that can be set through:
1. Configuration file
2. Environment variables
3. Command-line flags

## Configuration File

The default configuration file location is `~/.budgetassist/config.yaml`. You can specify a different location using the `--config` flag.

### Example Configuration

```yaml
# Database settings
database:
  driver: sqlite3
  path: ~/.budgetassist/data.db
  max_connections: 10
  timeout: 30s

# AI service configuration
ai:
  enabled: true
  model: gpt-4
  api_key: ${AI_API_KEY}  # Use environment variable
  timeout: 10s

# Logging configuration
logging:
  level: info
  format: json
  output: stdout

# Web server settings
server:
  port: 8080
  host: localhost
  timeout:
    read: 5s
    write: 10s
    idle: 60s

# Security settings
security:
  encryption_key: ${ENCRYPTION_KEY}
  allowed_origins:
    - http://localhost:3000
    - https://budgetassist.app
```

## Environment Variables

All configuration options can be set via environment variables using the prefix `BUDGET_ASSIST_`:

```bash
# Database
export BUDGET_ASSIST_DATABASE_DRIVER=sqlite3
export BUDGET_ASSIST_DATABASE_PATH=/path/to/db

# AI Service
export BUDGET_ASSIST_AI_ENABLED=true
export BUDGET_ASSIST_AI_API_KEY=your-api-key

# Logging
export BUDGET_ASSIST_LOGGING_LEVEL=debug

# Server
export BUDGET_ASSIST_SERVER_PORT=8080
```

## Command-line Configuration

Use the `config` command to view and modify settings:

```bash
# View all settings
budget-assist config list

# Get a specific setting
budget-assist config get database.path

# Set a value
budget-assist config set database.path /new/path/to/db

# Reset to default
budget-assist config reset database.path
```

## Configuration Options Reference

### Database Settings

| Option | Description | Default | Environment Variable |
|--------|-------------|---------|---------------------|
| `database.driver` | Database driver to use | sqlite3 | BUDGET_ASSIST_DATABASE_DRIVER |
| `database.path` | Path to database file | ~/.budgetassist/data.db | BUDGET_ASSIST_DATABASE_PATH |
| `database.max_connections` | Maximum number of connections | 10 | BUDGET_ASSIST_DATABASE_MAX_CONNECTIONS |
| `database.timeout` | Query timeout | 30s | BUDGET_ASSIST_DATABASE_TIMEOUT |

### AI Service Settings

| Option | Description | Default | Environment Variable |
|--------|-------------|---------|---------------------|
| `ai.enabled` | Enable AI features | true | BUDGET_ASSIST_AI_ENABLED |
| `ai.model` | AI model to use | gpt-4 | BUDGET_ASSIST_AI_MODEL |
| `ai.api_key` | API key for AI service | - | BUDGET_ASSIST_AI_API_KEY |
| `ai.timeout` | API call timeout | 10s | BUDGET_ASSIST_AI_TIMEOUT |

### Logging Settings

| Option | Description | Default | Environment Variable |
|--------|-------------|---------|---------------------|
| `logging.level` | Log level | info | BUDGET_ASSIST_LOGGING_LEVEL |
| `logging.format` | Log format (json/text) | json | BUDGET_ASSIST_LOGGING_FORMAT |
| `logging.output` | Log output destination | stdout | BUDGET_ASSIST_LOGGING_OUTPUT |

### Server Settings

| Option | Description | Default | Environment Variable |
|--------|-------------|---------|---------------------|
| `server.port` | HTTP server port | 8080 | BUDGET_ASSIST_SERVER_PORT |
| `server.host` | HTTP server host | localhost | BUDGET_ASSIST_SERVER_HOST |
| `server.timeout.read` | Read timeout | 5s | BUDGET_ASSIST_SERVER_TIMEOUT_READ |
| `server.timeout.write` | Write timeout | 10s | BUDGET_ASSIST_SERVER_TIMEOUT_WRITE |

## Advanced Configuration

### Using Multiple Profiles

Create profile-specific configurations:

```bash
# Create a profile
budget-assist config create-profile development

# Use a profile
budget-assist --profile development start

# Profile-specific config file
~/.budgetassist/config.development.yaml
```

### Configuration Validation

The application validates your configuration on startup. Common validation rules:

1. Port numbers must be between 1 and 65535
2. Timeouts must be positive durations
3. File paths must be valid and accessible
4. API keys must be properly formatted

### Secure Configuration

Best practices for secure configuration:

1. Use environment variables for sensitive values
2. Never commit API keys or secrets to version control
3. Use appropriate file permissions for config files
4. Regularly rotate sensitive credentials 