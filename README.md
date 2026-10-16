# gator

A command-line RSS feed aggregator written in Go and backed by PostgreSQL. Register users, add and follow RSS feeds, continuously scrape them for new posts, and browse the latest posts from the feeds you follow.

## Prerequisites

To install and run gator you'll need:

- **[Go](https://go.dev/doc/install)** (1.27 or newer), used to build and install the CLI
- **[PostgreSQL](https://www.postgresql.org/download/)** (15 or newer recommended), where users, feeds, and posts are stored

You'll also need [goose](https://github.com/pressly/goose) to run the database migrations:

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Installation

Install the `gator` CLI with `go install`:

```sh
go install github.com/mark-c-hall/gator@latest
```

This compiles a static binary and puts it in `$(go env GOPATH)/bin` (usually `~/go/bin`). Make sure that directory is on your `PATH`. After that you can run `gator` from anywhere, and you don't need the Go toolchain to run it.

> `go run .` is only for development. Use the installed `gator` binary for real use.

## Setup

### 1. Create the database

Start Postgres and create a database for gator:

```sh
psql -c "CREATE DATABASE gator;"
```

Your connection string will look something like:

```
postgres://username:password@localhost:5432/gator?sslmode=disable
```

### 2. Run the migrations

Clone the repo and apply the schema in `sql/schema` with goose:

```sh
git clone https://github.com/mark-c-hall/gator.git
cd gator
goose -dir sql/schema postgres "postgres://username:password@localhost:5432/gator?sslmode=disable" up
```

### 3. Create the config file

gator reads its config from `~/.gatorconfig.json`. Create it with your database URL:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable"
}
```

gator adds the `current_user_name` field itself when you register or log in.

## Usage

```sh
gator <command> [args...]
```

### Getting started

```sh
# Create a user and log in as them
gator register alice

# Add a feed (you automatically follow feeds you add)
gator addfeed "Hacker News" https://news.ycombinator.com/rss

# Start scraping feeds every minute (runs until you stop it with Ctrl+C)
gator agg 1m

# In another terminal, browse the 10 most recent posts from feeds you follow
gator browse 10
```

### Commands

| Command | Description |
| --- | --- |
| `register <name>` | Create a new user and log in as them |
| `login <name>` | Switch to an existing user |
| `users` | List all users (the current user is marked) |
| `addfeed <name> <url>` | Add a new feed and follow it |
| `feeds` | List all feeds in the database |
| `follow <url>` | Follow an existing feed |
| `following` | List the feeds the current user follows |
| `unfollow <url>` | Unfollow a feed |
| `agg <interval>` | Scrape feeds on a loop, e.g. `30s`, `1m`, `1h` |
| `browse [limit]` | Show the latest posts from followed feeds (default: 2) |
| `reset` | Delete all users and their data (useful for testing) |
