---
source: https://opencode.ai/docs/server/
fetched: 2026-01-08
title: Server
---

# OpenCode Server Documentation

## Overview

The OpenCode server is a headless HTTP service that exposes an OpenAPI endpoint, enabling programmatic interaction with OpenCode. When you run the standard `opencode` command, it launches both a TUI client and a backend server simultaneously.

## Core Features

### Architecture

"This architecture lets opencode support multiple clients and allows you to interact with opencode programmatically."

```
┌──────────────────────────────────────────────────────┐
│                    OpenCode                           │
├──────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │     TUI     │  │   IDE Ext   │  │   SDK App   │  │
│  │   Client    │  │   Client    │  │   Client    │  │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  │
│         │                │                │          │
│         └────────────────┼────────────────┘          │
│                          │                           │
│                   ┌──────▼──────┐                    │
│                   │   OpenCode  │                    │
│                   │   Server    │                    │
│                   │  (HTTP API) │                    │
│                   └─────────────┘                    │
└──────────────────────────────────────────────────────┘
```

### Server Startup

Launch a standalone server:

```bash
opencode serve
```

### Server Options

| Option | Default | Description |
|--------|---------|-------------|
| `--port` | `4096` | Server port |
| `--hostname` | `127.0.0.1` | Server hostname |
| `--mdns` | `false` | Enable mDNS discovery |
| `--cors` | - | CORS origins (comma-separated) |

### Configuration

```json
{
  "$schema": "https://opencode.ai/config.json",
  "server": {
    "port": 4096,
    "hostname": "127.0.0.1",
    "mdns": true,
    "cors": ["http://localhost:3000"]
  }
}
```

## API Categories

### Global

Health checks and event streaming:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/events` | GET | SSE event stream |

### Project/VCS

Project information and version control:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/project` | GET | Get current project |
| `/project/vcs` | GET | Version control info |

### Sessions

Session management:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/session` | GET | List sessions |
| `/session` | POST | Create session |
| `/session/:id` | GET | Get session |
| `/session/:id` | DELETE | Delete session |
| `/session/:id/share` | POST | Share session |

### Messages

Sending prompts and handling responses:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/session/:id/message` | POST | Send sync message |
| `/session/:id/prompt_async` | POST | Send async message |
| `/session/:id/command` | POST | Execute command |
| `/session/:id/shell` | POST | Execute shell |

### Files

Search and read operations:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/file/search` | POST | Search files |
| `/file/:path` | GET | Read file |
| `/file/:path/status` | GET | File status |

### Configuration

Provider and model configuration:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/config/provider` | POST | Set provider |
| `/config/model` | POST | Set model |

### Tools & LSP

Tool access and language server status:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/tool` | GET | List tools |
| `/tool/:name` | POST | Execute tool |
| `/lsp` | GET | LSP status |

### Agents

Available agents:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/agent` | GET | List agents |

### TUI Control

Remote terminal interface manipulation:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/tui/prompt` | POST | Prefill prompt |
| `/tui/toast` | POST | Show toast |
| `/tui/command` | POST | Execute command |
| `/tui/dialog` | POST | Open dialog |

## Notable Capabilities

### IDE Integration

The `/tui` endpoints enable external applications (like IDE plugins) to prefill prompts and drive the interface remotely.

### Async Operations

The `/session/:id/prompt_async` endpoint enables fire-and-forget operations, useful for background processing.

### OpenAPI Specification

Full OpenAPI 3.1 specification is available at the `/doc` endpoint for schema validation and client generation.

```bash
# Get OpenAPI spec
curl http://localhost:4096/doc
```

## Example Usage

### Start Server

```bash
opencode serve --port 4096 --hostname 0.0.0.0
```

### Health Check

```bash
curl http://localhost:4096/health
```

### Create Session

```bash
curl -X POST http://localhost:4096/session \
  -H "Content-Type: application/json" \
  -d '{"title": "My Session"}'
```

### Send Message

```bash
curl -X POST http://localhost:4096/session/abc123/message \
  -H "Content-Type: application/json" \
  -d '{"content": "Explain this codebase"}'
```
