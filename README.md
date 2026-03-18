# steamctl

A headless CLI tool that idles Steam games and unlocks achievements automatically, ordered from most common to rarest with randomized timing.

## Install

```bash
go install github.com/jeerrry/steamctl@latest
```

Or build from source:

```bash
git clone https://github.com/jeerrry/steamctl.git
cd steamctl
go build -o steamctl .
```

## Setup

1. Get a [Steam Web API key](https://steamcommunity.com/dev/apikey)
2. Find your 64-bit Steam ID at [steamid.io](https://steamid.io)
3. Create your config:

```bash
mkdir -p ~/.steamctl
cp config.example.toml ~/.steamctl/config.toml
# Edit with your API key and Steam ID
```

## Usage

```bash
# Add a game
steamctl add 3764200

# Preview the unlock schedule
steamctl dry-run

# Check progress
steamctl status

# Start the daemon
steamctl start

# Reset state for a game
steamctl reset 3764200
```

## Config

See [config.example.toml](config.example.toml) for all options.

```toml
[steam]
api_key = "YOUR_STEAM_WEB_API_KEY"
steam_id = "YOUR_STEAM_ID_HERE"

[[games]]
app_id = 3764200
name = "RE Requiem"
idle = true
unlock_achievements = true
time_range = "30h"
min_interval = "10m"
max_interval = "2h"
```

## Running as a Service

Copy the systemd unit file and enable it:

```bash
sudo cp steamctl.service /etc/systemd/system/steamctl@.service
sudo systemctl daemon-reload
sudo systemctl enable --now steamctl@$USER
```

Check logs:

```bash
journalctl -u steamctl@$USER -f
```

## Architecture

```
steamctl
├── cmd/           — CLI commands (cobra)
├── internal/
│   ├── config/    — TOML config parsing
│   ├── state/     — JSON state persistence
│   ├── steam/     — Steam Web API client
│   ├── sdk/       — Steamworks SDK bindings (planned)
│   └── scheduler/ — unlock scheduling logic
└── main.go
```

## Current Status

This is the initial version with all core logic implemented. Achievement unlocks are currently logged (stub) — actual Steamworks SDK integration via cgo will be added in a future phase.

## License

MIT
