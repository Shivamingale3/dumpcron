# Dumpcron

Scheduled database backup daemon for self-hosted servers. Backs up PostgreSQL, MySQL, and MongoDB databases on a daily schedule with streamed zstd compression. Emits journald events for PiDex Telegram alerts.

## Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/Shivamingale3/dumpcron/main/deploy/install.sh | sudo bash

# Edit the sample config with your databases
sudo nano /etc/dumpcron/config.yaml

# Validate
dumpcron validate

# Start
sudo systemctl enable --now dumpcron
```

**Prerequisites**: Linux (systemd), Go 1.22+ (if building from source), and the relevant dump tools for your databases.

## Commands

| Command | Description |
|---------|-------------|
| `dumpcron run` | Start the backup scheduler daemon |
| `dumpcron validate` | Validate config, dependencies, storage, and database connectivity |
| `dumpcron version` | Print version |
| `dumpcron uninstall` | Remove dumpcron, systemd service, and state |

## Supported Databases

| Database | Required Tool | File Extension |
|----------|--------------|----------------|
| PostgreSQL | `pg_dump` | `.sql.zst` |
| MySQL | `mysqldump` | `.sql.zst` |
| MongoDB | `mongodump` | `.json.zst` |

All dumps require `zstd` for compression.

## Configuration

Configuration file lives at `/etc/dumpcron/config.yaml`:

```yaml
backup_root: /srv/backups      # must exist and be writable
retention_days: 30             # backups older than this are deleted

jobs:
  - name: postgres_main
    type: postgres
    host: localhost
    port: 5432
    username: backup_user
    password: your_password
    databases:
      - app_db
      - auth_db
    time: "02:00"              # 24-hour format, daily

  - name: mysql_main
    type: mysql
    host: localhost
    port: 3306
    username: backup_user
    password: your_password
    databases:
      - users_db
    time: "03:00"

  - name: mongo_main
    type: mongo
    host: localhost
    port: 27017
    username: backup_user
    password: your_password
    databases:
      - analytics
    time: "04:00"
```

Changes require a restart: `sudo systemctl restart dumpcron`

## Backup Output Structure

```
/srv/backups/
├── postgres/
│   ├── app_db_2026-06-07_02-00.sql.zst
│   └── auth_db_2026-06-07_02-00.sql.zst
├── mysql/
│   └── users_2026-06-07_03-00.sql.zst
└── mongo/
    └── analytics_2026-06-07_04-00.json.zst
```

## Logs

```bash
journalctl -u dumpcron -f
```

## PiDex Integration

Dumpcron emits 9 event types to journald. Copy the PiDex config and register:

```bash
sudo pidex setup          # → Manage custom services → dumpcron.conf → Register
sudo systemctl restart pidex
```

| Event | Severity | When |
|-------|----------|------|
| `DUMPCRON_STARTED` | INFO | Service started |
| `DUMPCRON_STOPPED` | WARNING | Service stopped |
| `DUMPCRON_CONFIG_INVALID` | CRITICAL | Config validation failed |
| `DUMPCRON_DEPENDENCY_MISSING` | CRITICAL | Required binary not found |
| `DUMPCRON_STORAGE_INVALID` | CRITICAL | Backup directory unavailable |
| `BACKUP_JOB_STARTED` | INFO | Job execution started |
| `BACKUP_JOB_COMPLETED` | INFO | Job completed successfully |
| `BACKUP_JOB_FAILED` | CRITICAL | Job completed with failures |
| `DATABASE_BACKUP_FAILED` | CRITICAL | Individual database dump failed |

## Building from Source

```bash
git clone https://github.com/Shivamingale3/dumpcron.git
cd dumpcron
go build -ldflags="-s -w" -o dumpcron ./cmd/dumpcron

sudo install -m 755 dumpcron /usr/local/bin/dumpcron
sudo mkdir -p /etc/dumpcron /var/lib/dumpcron
sudo cp config/dumpcron.conf /etc/pidex/custom.d/dumpcron.conf  # optional
sudo cp deploy/dumpcron.service /etc/systemd/system/
sudo systemctl daemon-reload
```

## Running Tests

```bash
go test ./...
```

## Uninstall

```bash
sudo dumpcron uninstall
```

Or manually:

```bash
sudo systemctl disable --now dumpcron
sudo rm -f /usr/local/bin/dumpcron /etc/systemd/system/dumpcron.service
sudo rm -rf /etc/dumpcron /var/lib/dumpcron
sudo rm -f /etc/pidex/custom.d/dumpcron.conf
sudo systemctl daemon-reload
```
