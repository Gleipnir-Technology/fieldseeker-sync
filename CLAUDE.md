# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Top-level guidance

Any time Claude produces code it should be within a git branch. Create a branch name based on the initial request and make all commits to that branch. The goal is to produce a large number of small commits. Each commit starts with a single line in imperative style indicating what the change is. There is then a blank line, then however many lines necessary to explain why the commit is necessary. Expect the commit to be reviewed by a competent senior engineer that understands the mechanics of how it works and is looking to understand the thought process for why a particular approach was chosen. Ideal commits are quite small and fully self-contained.

After a particularly productive session Claude recommends changes to CLAUDE.md itself to improve and refine instructions for future coding agents.

## Development Environment Setup

Use Nix shell for development environment:
```bash
nix-shell
```

This provides Go, PostgreSQL, Goose (migrations), Ninja (build), and other development tools.

## Database Operations

Start development database:
```bash
./start-database.sh
```

Check migration status:
```bash
env GOOSE_DRIVER=postgres GOOSE_DBSTRING="user=fieldseeker dbname=fieldseeker password=letmein" goose status
```

Connection string for development:
```bash
env DATABASE_URL=postgresql://fieldseeker:letmein@localhost:5432 ./fieldseeker-sync
```

In general Claude should not do any database operations directly. It's just useful to be aware of these things to answer questions and make suggestions.

## Build Commands

Build using Ninja (preferred):
```bash
ninja
```

Build manually:
```bash
go build
```

Build specific binaries:
```bash
ninja bin/webserver
ninja bin/autobuild
ninja bin/export
```

## Testing

Run tests with race detection:
```bash
go test -race -count=1 -timeout=30s
```

Run tests on specific packages:
```bash
go test -race -count=1 -timeout=30s ./cmd/webserver
```

Claude should not run tests directly, but may make suggestions for new tests to add, or for tests to run.

## Code Quality

Format code (handled by lefthook pre-commit):
```bash
gofmt -w .
```

## Architecture Overview

This is a FieldSeeker synchronization bridge that connects a PostgreSQL database with FieldSeeker via ArcGIS API. The application consists of multiple binaries:

- **webserver**: HTTP server with authentication, HTML templates, and REST API endpoints
- **export**: Synchronization process that exports data from FieldSeeker to local database
- **autobuild**: Development tool with file watching and terminal UI
- **registration**: User registration system
- **login**: Authentication system
- **dump**: Data export utility
- **schema**: Schema management utility

### Key Components

**Database Layer** (`database.go`):
- PostgreSQL connection pooling via pgx
- Embedded migrations system
- Tables prefixed with `FS_` are FieldSeeker tables (read-only, managed by sync process)
- History tables (`History_*`) track all changes with versioning
- Each `FS_` table has an "updated" column for incremental sync

**Web Server** (`cmd/webserver/`):
- Chi router with middleware stack
- Session-based authentication using SCS
- Content negotiation between JSON and HTML
- Template-based HTML rendering
- REST API endpoints under `/api/`

**Configuration** (`config.go`):
- Uses Viper for TOML configuration
- Searches `/etc/`, `$HOME/.config`, and current directory
- Supports ArcGIS, database, and webhook configuration

**Types** (`types.go`):
- Core data structures for FieldSeeker entities
- Geometry handling for spatial data
- Time parsing utilities for FieldSeeker timestamps

### Authentication Flow

The web server uses session-based authentication with the following flow:
1. Unauthenticated requests to protected endpoints redirect to `/login`
2. API requests return 401 with "Login required" message
3. Sessions store `display_name` and `username`
4. Content negotiation determines response format (HTML vs JSON)

### Database Schema

- **FS_*** tables: Read-only FieldSeeker data synchronized from ArcGIS
- **History_*** tables: Versioned history of all changes to FS_ tables
- **User management**: Handled through separate registration/login systems
- **Audit trail**: Tracked via migrations 00005 and 00006

### Development Workflow

1. Use `./start-database.sh` to start PostgreSQL in Docker
2. Run migrations automatically via `InitDB()` 
3. Use `ninja` for builds or `go build` for simple compilation
4. Run `cmd/webserver/main.go` for development server on port 3000
5. Use `cmd/autobuild/main.go` for file watching during development

## Git Hooks

Lefthook manages git hooks:
- **pre-commit**: Runs `gofmt -w` on staged Go files
- **pre-push**: Runs `go test -race -count=1 -timeout=30s`
