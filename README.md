# Stockyard Roundup

**Meeting notes and decisions log — structured notes, action items, decision log, search**

Part of the [Stockyard](https://stockyard.dev) family of self-hosted developer tools.

## Quick Start

```bash
docker run -p 9170:9170 -v roundup_data:/data ghcr.io/stockyard-dev/stockyard-roundup
```

Or with docker-compose:

```bash
docker-compose up -d
```

Open `http://localhost:9170` in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9170` | HTTP port |
| `DATA_DIR` | `./data` | SQLite database directory |
| `ROUNDUP_LICENSE_KEY` | *(empty)* | Pro license key |

## Free vs Pro

| | Free | Pro |
|-|------|-----|
| Limits | 20 meetings | Unlimited meetings and decisions |
| Price | Free | $2.99/mo |

Get a Pro license at [stockyard.dev/tools/](https://stockyard.dev/tools/).

## Category

Operations & Teams

## License

Apache 2.0
