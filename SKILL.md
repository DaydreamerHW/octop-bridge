---
name: octop-bridge
description: Delegate tasks and chat with Octop AI agents on NAS.
version: 0.1.0
author: Xia Qian, Hermes Agent
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [octop, bridge, multi-agent, nas, agent-to-agent]
    category: autonomous-ai-agents
---

# Octop Bridge Skill

Interact, chat, and delegate specialized tasks to self-hosted Octop AI agents (like 小通) running on the local NAS via WebSocket & REST API.

## When to Use

Trigger this skill whenever:
- The user asks to query or delegate a task to "小通" (Xiao Tong), "Octop", or the "NAS agent":
  - *"问下小通..."*, *"让小通帮我查..."*, *"用Octop处理..."*
- Delegating domestic Chinese ecosystem tasks powered by Octop's specialized MCP connectors:
  - **Travel & Tickets**: Flights, train tickets, and hotel searches (Ctrip / 携程问道, Fliggy / 飞猪).
  - **Transport & Navigation**: Didi routes, maps, and local POI lookups (Didi / 滴滴).
  - **News & Media**: Real-time domestic news (Tencent News / 腾讯新闻).
  - **Channel Delivery**: Pushing messages or files to the user's WeChat (微信) channel through Octop.
  - **NAS Knowledge Base**: Querying RAG documents stored in the Octop knowledge base on NAS.
- Do NOT use for: Code debugging, local file management, and deep reasoning that Hermes can handle directly.

## Prerequisites

- Octop service running on NAS (default: `192.168.5.192:8089`).
- Python 3 with `websockets` library installed.
- Bridge script at `~/Develop/skills/octop-bridge/scripts/octop_bridge.py`.

## How to Run

Execute the bridge script via the `terminal` tool:

```bash
# Basic single query (streams response to stdout)
python3 ~/Develop/skills/octop-bridge/scripts/octop_bridge.py "<your query or instruction>"

# Programmatic invocation with JSON output (captures thread_id and tools called)
python3 ~/Develop/skills/octop-bridge/scripts/octop_bridge.py "<query>" --json

# Resume/continue an existing conversation thread
python3 ~/Develop/skills/octop-bridge/scripts/octop_bridge.py "<follow-up>" --thread-id "<thread_id>"

# Target a different agent ID or specify timeout
python3 ~/Develop/skills/octop-bridge/scripts/octop_bridge.py "<query>" --agent-id "PQG34P" --timeout 180
```

## Quick Reference

| Flag | Purpose | Default |
| :--- | :--- | :--- |
| `query` / `-q` | Text prompt/task to send to Octop agent | Required |
| `--agent-id` / `-a` | Target Octop agent ID | `PQG34P` (小通 · 通用助手) |
| `--thread-id` / `-t` | Existing thread ID to continue conversation | None (reuses dashboard thread) |
| `--host` | Octop server IP or hostname | `192.168.5.192` (or `OCTOP_HOST`) |
| `--port` | Octop HTTP/WS port | `8089` (or `OCTOP_PORT`) |
| `--mode` / `-m` | Turn permission mode (`craft`, `ask`, `plan`) | `craft` |
| `--json` | Return structured JSON to stdout | Disabled |
| `--timeout` | Maximum wait time in seconds | `300` |

## Procedure

1. **Formulate the Delegated Task**:
   - Construct a clear, self-contained prompt specifying all needed criteria (departure/arrival cities, exact dates, constraints, preferred output format).
2. **Execute Bridge Command**:
   - Call the `terminal` tool to run `octop_bridge.py`:
     ```bash
     python3 ~/Develop/skills/octop-bridge/scripts/octop_bridge.py "<formulated_prompt>" --json
     ```
3. **Parse & Present the Outcome**:
   - Extract the `response` and `tools_called` from the output.
   - If Octop invoked travel or MCP tools, report the factual findings directly to the user.
   - If a continuous multi-turn dialogue is required, retain the returned `thread_id` for subsequent turns.

## Pitfalls

- **Network reachability**: Octop runs on the local LAN (`192.168.5.192:8089`). If unreachable, verify the NAS host is online and Octop service is active.
- **Complex tool latency**: Octop tasks involving multi-step MCP tools (e.g. searching flights and comparing airlines) may take 30–60 seconds; set `--timeout 180` or higher to avoid premature timeouts.
- **Credential security**: Default credentials are read from environment variables (`OCTOP_PASSWORD`, `OCTOP_USERNAME`) or safe local defaults. Avoid echoing passwords in chat logs.

## Verification

- The script exits with code 0.
- JSON output returns `"status": "success"` and non-empty `"response"`.
