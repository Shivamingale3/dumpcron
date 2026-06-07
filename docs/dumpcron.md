# Dumpcron v1 - Technical Design Specification

## Overview

Dumpcron is a lightweight Linux service that performs scheduled database backups for locally reachable databases.

The application is designed primarily for self-hosted servers, Raspberry Pi servers, homelabs, VPSs, and small deployments.

Dumpcron is:

* Config-file driven
* Stateless
* Systemd-native
* Queue-based
* Event emitting
* PiDex compatible

Dumpcron is NOT:

* A web application
* A remote backup controller
* A multi-server management platform
* A backup storage solution
* A database restore manager

Its sole responsibility is:

1. Validate configuration
2. Execute scheduled backups
3. Compress backups
4. Delete expired backups
5. Emit operational events

---

# Supported Databases

Version 1 supports:

* PostgreSQL
* MySQL
* MongoDB

No other database engines are supported.

---

# Configuration Location

Single configuration file:

```text
/etc/dumpcron/config.yaml
```

No includes.

No multiple config files.

No config directories.

No hot reload.

Configuration changes require service restart.

---

# Example Configuration

```yaml
backup_root: /srv/backups

retention_days: 30

jobs:
  - name: postgres_main

    type: postgres

    host: localhost
    port: 5432

    username: backup_user
    password: super_secret

    databases:
      - app_db
      - auth_db

    time: "02:00"

  - name: mysql_main

    type: mysql

    host: localhost
    port: 3306

    username: backup_user
    password: another_secret

    databases:
      - users_db

    time: "03:00"

  - name: mongo_main

    type: mongo

    host: localhost
    port: 27017

    username: backup_user
    password: mongo_secret

    databases:
      - analytics

    time: "04:00"
```

---

# Configuration Rules

## Global Fields

### backup_root

Root directory where backups are stored.

Example:

```yaml
backup_root: /srv/backups
```

Must exist.

Must be writable.

---

### retention_days

Number of days to keep backups.

Example:

```yaml
retention_days: 30
```

Backups older than this value are automatically deleted.

---

# Job Fields

## name

Human-readable unique identifier.

Example:

```yaml
name: postgres_main
```

Must be unique.

---

## type

Supported values:

```yaml
type: postgres
type: mysql
type: mongo
```

---

## host

Database hostname.

Example:

```yaml
host: localhost
```

---

## port

Database port.

Examples:

```yaml
port: 5432
port: 3306
port: 27017
```

---

## username

Database login username.

---

## password

Database login password.

Stored in plain text.

Version 1 does not support:

* Encryption
* Vault
* Environment variables
* Secret managers

---

## databases

List of databases to backup.

Example:

```yaml
databases:
  - app_db
  - auth_db
```

Must contain at least one entry.

---

## time

Daily execution time.

Format:

```yaml
time: "02:00"
```

24-hour format only.

Dumpcron only supports daily schedules.

Not supported:

* Hourly
* Weekly
* Monthly
* Cron expressions

---

# CLI Commands

## Validate

```bash
dumpcron validate
```

Performs complete validation.

Validation succeeds only if every configured job is valid.

---

# Validation Checks

## Config Validation

Validate:

* YAML syntax
* Required fields
* Unique job names
* Supported database types
* Valid time format
* Positive retention value

---

## Dependency Validation

Only validate dependencies required by configured jobs.

### PostgreSQL

Required:

```bash
pg_dump
```

---

### MySQL

Required:

```bash
mysqldump
```

---

### MongoDB

Required:

```bash
mongodump
```

---

### Compression

Required:

```bash
zstd
```

---

## Storage Validation

Validate:

* backup_root exists
* backup_root writable

---

## Database Validation

For every configured job:

Validate:

* Host reachable
* Port reachable
* Authentication successful
* Selected databases exist

---

# Validation Success

Example:

```text
✓ Configuration valid
✓ Dependencies present
✓ Storage valid
✓ Database connectivity valid

Configuration OK
```

---

# Validation Failure

Example:

```text
✗ pg_dump missing

✗ Job postgres_main
Authentication failed

Validation failed
```

Exit code must be non-zero.

---

# Service

Systemd service:

```text
dumpcron.service
```

Example:

```bash
sudo systemctl enable dumpcron
sudo systemctl start dumpcron
```

---

# Startup Flow

When service starts:

```text
Load Config
      ↓
Validate Config
      ↓
Validate Dependencies
      ↓
Validate Storage
      ↓
Validate Databases
      ↓
Start Scheduler
```

If any validation fails:

```text
Emit event
Exit process
```

Scheduler never starts.

---

# Scheduler Design

Dumpcron uses a single scheduler thread.

Daily schedule only.

Every minute:

```text
Check current time
Find due jobs
Queue due jobs
```

---

# Queue Design

Single queue.

Single worker.

No parallel execution.

Example:

```text
02:00 Job A
02:00 Job B
02:00 Job C
```

Execution:

```text
Job A
Job B
Job C
```

Sequentially.

---

# Long Running Jobs

If a job execution overlaps the next day's execution:

Example:

```text
Run #1 still running
Run #2 becomes due
```

Behavior:

```text
Run #2 enters queue
```

Nothing is skipped.

Nothing is cancelled.

---

# Backup Process

For each database:

```text
Connect
Dump
Compress
Store
```

Executed independently.

---

# PostgreSQL

Command concept:

```bash
pg_dump database_name
```

---

# MySQL

Command concept:

```bash
mysqldump database_name
```

---

# MongoDB

Command concept:

```bash
mongodump database_name
```

---

# Compression

Compression must be streamed.

Never create raw dump files first.

Correct:

```bash
pg_dump app_db | zstd > backup.zst
```

Incorrect:

```bash
pg_dump > dump.sql
zstd dump.sql
```

---

# Output Structure

Directory structure:

```text
/srv/backups

├── postgres
├── mysql
└── mongo
```

Automatically created.

---

# PostgreSQL Files

```text
/srv/backups/postgres/

app_db_2026-06-07_02-00.sql.zst
auth_db_2026-06-07_02-00.sql.zst
```

---

# MySQL Files

```text
/srv/backups/mysql/

users_2026-06-07_03-00.sql.zst
```

---

# Mongo Files

```text
/srv/backups/mongo/

analytics_2026-06-07_04-00.json.zst
```

---

# Partial Failures

Example:

```text
app_db       success
auth_db      success
inventory_db failed
```

Result:

```text
Completed with failures
```

Successful backups remain.

Nothing is deleted.

Nothing is rolled back.

---

# Retention

After successful backup processing:

```text
Delete files older than retention_days
```

Default:

```yaml
retention_days: 30
```

Applies independently to:

```text
postgres/
mysql/
mongo/
```

directories.

---

# Logging

Dumpcron logs to journald.

Systemd captures stdout/stderr.

---

# PiDex Integration

Dumpcron integrates through journald messages.

No API.

No socket.

No webhook.

No SDK.

---

# Events

## Service Events

```text
DUMPCRON_STARTED
DUMPCRON_STOPPED
```

---

## Validation Events

```text
DUMPCRON_CONFIG_INVALID
DUMPCRON_DEPENDENCY_MISSING
DUMPCRON_STORAGE_INVALID
```

---

## Job Events

```text
BACKUP_JOB_STARTED
BACKUP_JOB_COMPLETED
BACKUP_JOB_FAILED
```

---

## Database Events

```text
DATABASE_BACKUP_FAILED
```

Database success events are intentionally omitted to avoid notification spam.

---

# Failure Handling

## Missing Dependency

Example:

```text
pg_dump missing
```

Action:

```text
Emit DUMPCRON_DEPENDENCY_MISSING
Exit
```

---

## Invalid Config

Action:

```text
Emit DUMPCRON_CONFIG_INVALID
Exit
```

---

## Storage Unavailable

Action:

```text
Emit DUMPCRON_STORAGE_INVALID
Exit
```

---

## Database Authentication Failure

Action:

```text
Emit DUMPCRON_CONFIG_INVALID
Exit
```

---

## Runtime Backup Failure

Action:

```text
Emit DATABASE_BACKUP_FAILED
Continue remaining databases
```

---

# State Management

Dumpcron is completely stateless.

No SQLite.

No PostgreSQL.

No internal database.

No persistent queue.

State exists only in:

```text
Configuration file
Filesystem backups
In-memory queue
```

---

# Project Structure

```text
dumpcron/

├── cli/
│   └── validate.py
│
├── config/
│   └── loader.py
│
├── validation/
│   ├── config.py
│   ├── dependencies.py
│   ├── storage.py
│   └── databases.py
│
├── scheduler/
│   ├── scheduler.py
│   ├── queue.py
│   └── worker.py
│
├── drivers/
│   ├── postgres.py
│   ├── mysql.py
│   └── mongo.py
│
├── retention/
│   └── cleanup.py
│
├── events/
│   └── journald.py
│
├── daemon.py
│
└── main.py
```

---

# Development Order

Phase 1

* Config parser
* Validation framework

Phase 2

* PostgreSQL backup driver

Phase 3

* MySQL backup driver

Phase 4

* Mongo backup driver

Phase 5

* Scheduler

Phase 6

* Queue worker

Phase 7

* Retention

Phase 8

* Journald event emission

Phase 9

* Systemd packaging

Phase 10

* End-to-end testing

```

Dumpcron v1 is complete when all phases above are implemented and verified on a Linux system.
```
