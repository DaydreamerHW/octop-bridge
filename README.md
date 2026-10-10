# Octop Bridge 🐙

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg?logo=go)](https://go.dev)
[![Agent Skills Compatible](https://img.shields.io/badge/AgentSkills-Compatible-success.svg)](https://agentskills.io)

**Octop Bridge** is a universal client and AgentSkill bridge to chat, collaborate, and delegate tasks to self-hosted [Octop](https://github.com/TencentCloud/Octop) AI agents (such as "小通" / General Assistant) running on home NAS, Linux servers, or cloud instances.

It connects autonomous coding agents (like **Hermes Agent**, **Claude Code**, **OpenCode**, and **Codex**) with Octop's native domestic ecosystem and MCP tools (e.g. Ctrip/携程, Fliggy/飞猪, Didi/滴滴, Tencent News, WeChat IM channels, and NAS knowledge bases).

---

## 🌟 Key Features

- **⚡ Full-Duplex WebSocket Streaming**: Connects directly to `/api/agents/{agent_id}/chat/ws` for real-time typewriter token output and tool execution status tracking.
- **🔐 Automatic Authentication & Token Caching**: Seamlessly logs in with credentials, caches JWT bearer tokens with expiration awareness, and minimizes round-trip latency.
- **🧭 Dynamic Agent Discovery & Fuzzy Routing**: No need to memorize cryptic internal IDs. Use `list-agents` to inspect available agents, or chat using natural names (e.g. `--agent "小通"`).
- **🔄 Session Scoping (1:1 Binding)**: Automatically detects caller session (`HERMES_SESSION_ID`) and creates a dedicated, clean thread on first contact, reusing it across subsequent turns. Prevents context bleeding across tasks. Pass `--new` to force a reset.
- **🌐 Network & Proxy Resilient**: Supports HTTP, HTTPS, WS, WSS, custom ports, subpaths, and `--insecure` mode for self-signed certificates on home NAS.
- **📦 Clean Architecture & Standards Compliant**: Ready to install as a standalone Python CLI or load directly into any AgentSkills-compliant framework via `SKILL.md`.

---

## 🏗️ Architecture

```
┌─────────────────────────────────┐
│   Hermes Agent / Claude Code    │  (Coding, System Ops, Dev Tasks)
└───────────────┬─────────────────┘
                │
                ▼   via octop_bridge (CLI / Python client)
┌────────────────────────────────────────────────────────┐
│   WebSocket / REST API (:8089)                         │
└───────────────┬────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────┐
│    Octop Runtime on NAS/Linux   │  (Xiao Tong / Specialists)
├─────────────────────────────────┤
│  • Ctrip / Fliggy (Travel/Tickets)
│  • Didi (Maps & Routing)        │
│  • Tencent News & Knowledge RAG │
│  • WeChat / Feishu IM Channels  │
└─────────────────────────────────┘
```

---

## 🚀 Quick Start

### 1. Installation

```bash
# Clone the repository
git clone https://github.com/DaydreamerHW/octop-bridge.git
cd octop-bridge

# Build binary
go build -o octop-bridge .

# Move to your PATH (optional)
sudo mv octop-bridge /usr/local/bin/
# or install directly with go
go install github.com/DaydreamerHW/octop-bridge@latest
```

### 2. Configuration

You can configure connection parameters via environment variables or an interactive setup wizard:

```bash
# Option A: Interactive setup wizard (saved to ~/.config/octop-bridge/config.json)
octop-bridge setup

# Option B: Environment variables
export OCTOP_URL="http://192.168.5.192:8089"
export OCTOP_USERNAME="your_username"
export OCTOP_PASSWORD="your_password"
# export OCTOP_AGENT_ID="PQG34P" # Optional default
```

### 3. Usage Examples

```bash
# 1. Discover all agents on the instance
octop-bridge list-agents

# 2. Chat with the default agent
octop-bridge "What is the capital of Yunnan?"

# 3. Target a specific agent by display name
octop-bridge --agent "小通" "Check the cheapest flights from Kunming to Wuhan tomorrow"

# 4. Programmatic JSON output (captures thread_id and tools called)
octop-bridge "Find today's top tech news" --json

# 5. Continue a conversation thread
octop-bridge "Can you recommend the morning flight among those?" --thread-id "thr_01M42Y9CWQ2MVW6S3A7QVS4FHQ"

# 6. Manage agent runtime & lifecycle
octop-bridge agent list                         # List all agents
octop-bridge agent get "小通"                   # Get agent details
octop-bridge agent status "小通"                # Inspect real-time status, channels & errors
octop-bridge agent reload "小通"                # Hot-reload agent runtime, skills & connectors
octop-bridge agent start "小通"                 # Start agent runtime
octop-bridge agent stop "小通"                  # Stop agent runtime
octop-bridge agent delete "test-agent" --yes    # Delete agent and workspace

# 7. Manage and inspect Octop skills
octop-bridge skill list                         # List all skills for active agent
octop-bridge skill list --kind workspace       # List only workspace installed skills
octop-bridge skill disable "gemini-image-gen"   # Disable a skill remotely
octop-bridge skill enable "gemini-image-gen"    # Enable a skill remotely
octop-bridge skill install ./path/to/my-skill   # Deploy local skill directory/zip to Octop in one command!
octop-bridge skill remove "my-skill"            # Delete an installed skill from agent workspace
```

---

## ⚙️ Configuration Reference

| Parameter | Environment Variable | Config File Key | Description | Default |
| :--- | :--- | :--- | :--- | :--- |
| `--url` | `OCTOP_URL` | `base_url` | Full HTTP/HTTPS URL of Octop instance | `http://192.168.5.192:8089` |
| `--username` / `-u` | `OCTOP_USERNAME` | `username` | Octop login username | `None` |
| `--password` / `-p` | `OCTOP_PASSWORD` | `password` | Octop login password | `None` |
| `--token` | `OCTOP_TOKEN` | `token` | Direct JWT bearer token (skips login) | `None` |
| `--agent` / `-a` | `OCTOP_AGENT_ID` | `default_agent_id` | Target Agent ID or fuzzy name | First available / `PQG34P` |
| `--insecure` / `-k` | `OCTOP_INSECURE` | `insecure` | Allow unverified TLS certificates | `false` |
| `--timeout` | `OCTOP_TIMEOUT` | `timeout` | Execution timeout in seconds | `180` |

---

## 🤖 Using as an Agent Skill

This repository includes a standardized `SKILL.md` in the root folder.

### For Hermes Agent:
Copy or symlink `octop-bridge` into your Hermes skills directory:
```bash
ln -s "$(pwd)" ~/.hermes/skills/autonomous-ai-agents/octop-bridge
```
Then simply ask Hermes in chat:
> *"让小通帮我查一下后天去上海的机票"*

Hermes will automatically detect and invoke the skill, delegate the query to Octop, and report the real-time findings.

---

## 🧪 Testing

Run the test suite:
```bash
go test -v ./...
```

---

## 📄 License

Released under the [MIT License](LICENSE).
