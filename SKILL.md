---
name: octop-bridge
description: Delegate tasks and chat with Octop AI agents on NAS.
version: 1.0.0
author: Xia Qian, Hermes Agent
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [octop, bridge, multi-agent, nas, agent-to-agent]
    category: autonomous-ai-agents
---

# Octop Bridge Skill

Interact, chat, and delegate specialized tasks to self-hosted Octop AI agents (like 小通) running on the local NAS via WebSocket & REST API using the fast Go-powered CLI (`octop-bridge`).

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
- Discovering or managing Octop agents & threads:
  - *"看看Octop有哪些专家"*, *"列出小通的对话记录"*
- Do NOT use for: Code debugging, local file management, and deep reasoning that Hermes can handle directly.

## Prerequisites

- Octop service running on NAS (default: `192.168.5.192:8089`).
- `octop-bridge` CLI installed in PATH (e.g. `/Users/xiaqian/.local/bin/octop-bridge`).
- Configuration saved in `~/.config/octop-bridge/config.json`.

## How to Run

Execute `octop-bridge` via the `terminal` tool:

```bash
# Basic single query (streams response to stdout)
octop-bridge "<your query or instruction>"

# Programmatic invocation with JSON output (captures thread_id, agent_name, and tools)
octop-bridge "<query>" --json

# Start a fresh isolated conversation session (new thread)
octop-bridge --new "<query>" --json

# Resume/continue an existing conversation thread
octop-bridge "<follow-up>" --thread-id "<thread_id>"

# Target a different agent by display name or ID
octop-bridge --agent "小通" "<query>"

# List all available agents or recent conversation threads
octop-bridge list-agents --json
octop-bridge list-threads --json

# Manage Octop skills remotely
octop-bridge skill list --json
octop-bridge skill enable <skill-name>
octop-bridge skill disable <skill-name>
octop-bridge skill install <local-skill-directory-or-zip>
octop-bridge skill remove <skill-name>
```

## Quick Reference

| Flag | Purpose | Default |
| :--- | :--- | :--- |
| `query` | Text prompt/task to send to Octop agent | Required |
| `--agent` / `-a` | Target Agent ID or fuzzy name | `PQG34P` (小通 · 通用助手) |
| `--new` / `-n` | Start a brand new conversation thread | `false` |
| `--thread-id` / `-t` | Existing thread ID to continue conversation | None (reuses active thread) |
| `--mode` / `-m` | Turn permission mode (`craft`, `ask`, `plan`) | `craft` |
| `--json` | Return structured JSON to stdout | Disabled |
| `--timeout` | Maximum wait time in seconds | `180` |

## Procedure

1. **Formulate the Delegated Task**:
   - Construct a clear, self-contained prompt specifying all needed criteria (cities, exact dates, constraints).
2. **Execute Bridge Command**:
   - Call the `terminal` tool:
     ```bash
     octop-bridge "<formulated_prompt>" --json
     ```
3. **Parse & Present the Outcome**:
   - Extract the `response` and `tools_called` from the output.
   - If Octop invoked travel or MCP tools, report the factual findings directly to the user.
   - If a continuous multi-turn dialogue is required, retain the returned `thread_id` for subsequent turns.

## Pitfalls

- **Complex tool latency**: Octop tasks involving multi-step MCP tools (e.g. searching flights and comparing airlines) may take 20–50 seconds; allow sufficient timeout.
- **Credential security**: Credentials and tokens are managed via `~/.config/octop-bridge/config.json` and cached in `~/.cache/octop-bridge/token_cache.json`. Never echo credentials into chat logs.

## Verification

- The command exits with code 0.
- JSON output returns `"status": "success"` and non-empty `"response"`.
