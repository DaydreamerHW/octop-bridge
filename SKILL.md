---
name: octop-bridge
description: Delegate tasks, manage expert teams, and chat with Octop AI agents on NAS.
version: 1.1.0
author: Xia Qian, Hermes Agent
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [octop, bridge, multi-agent, expert-teams, nas, agent-to-agent]
    category: autonomous-ai-agents
---

# Octop Bridge Skill

Interact, chat, manage expert teams (专家组), and delegate specialized tasks to self-hosted Octop AI agents (like 小通, 老钱, or composite expert teams) running on the local NAS via WebSocket & REST API using the fast Go-powered CLI (`octop-bridge`).

## When to Use

Trigger this skill whenever:
- The user asks to query or delegate a task to "小通" (Xiao Tong), "Octop", or specific NAS agents/teams:
  - *"问下小通..."*, *"让小通帮我查..."*, *"用Octop处理..."*, *"让专家组会诊..."*
- Managing or delegating to **Expert Teams (专家组)**:
  - *"创建一个专家组..."*, *"列出Octop上的团队/专家组"*, *"往专家组里加一个成员"*
  - Delegating complex multi-domain workflows to an Octop Team Host (dispatches to specialists via `ask_agent`).
- Browsing expert templates or instantiating new specialist agents:
  - *"看看Octop有哪些内置专家模板"*, *"从模板创建一个运维工程师/热点记者专家"*
- Delegating domestic Chinese ecosystem tasks powered by Octop's specialized MCP connectors:
  - **Travel & Tickets**: Flights, train tickets, and hotel searches (Ctrip / 携程问道, Fliggy / 飞猪).
  - **Transport & Navigation**: Didi routes, maps, and local POI lookups (Didi / 滴滴).
  - **News & Media**: Real-time domestic news (Tencent News / 腾讯新闻, Weibo, Zhihu trends).
  - **Channel Delivery**: Pushing messages or files to the user's WeChat (微信) channel through Octop.
  - **NAS Knowledge Base**: Querying RAG documents stored in the Octop knowledge base on NAS.
- Do NOT use for: Code debugging, local file management, and deep reasoning that Hermes can handle directly.

## Prerequisites

- Octop service running on NAS (default: `192.168.5.192:8089`).
- `octop-bridge` CLI installed in PATH (`~/.local/bin/octop-bridge`).
- Configuration saved in `~/.config/octop-bridge/config.json`.

## Session Scoping (1:1 Binding)

- **Default Behavior**: When called from a Hermes session, `octop-bridge` automatically detects the caller's session ID (`HERMES_SESSION_ID`) and creates a **dedicated, clean conversation thread** on first contact with an agent or team host.
- **Continuity**: Follow-up calls within the same Hermes session automatically resume and continue that target's dedicated thread, preserving context and memory while preventing cross-session context bleeding.
- **Isolation**: Each Hermes session is completely isolated from other sessions and never pollutes the user's web dashboard thread.
- **Force Reset**: Pass `--new` (`-n`) to force a fresh thread within the current session when explicitly requested by the user.

## Command Reference

### 1. Chat & Delegation
```bash
# Chat with default agent (小通)
octop-bridge "<query>" --json

# Chat with a specific specialist or expert team by name/ID
octop-bridge --agent "综合服务组" "<query>" --json
octop-bridge --agent "老钱" "分析一下当前上证指数走势" --json

# Force a clean conversation thread
octop-bridge --new --agent "<agent-or-team>" "<query>" --json

# Resume explicit thread
octop-bridge --thread-id "<thread_id>" "<follow-up>"
```

### 2. Expert Teams Management (专家组)
```bash
# List all expert teams
octop-bridge list-teams --json
octop-bridge team list

# Get details and roster of an expert team
octop-bridge team get "<team-id-or-name>"

# Create a new expert team with members (IDs or fuzzy names)
octop-bridge team create --name "投研生活组" --description "整合证券分析与出行生活" --members "PQG34P,9TS55H"

# Update an expert team's metadata or member roster
octop-bridge team update "投研生活组" --members "PQG34P,9TS55H,FD4Z2A" --description "新增热点分析"

# Delete an expert team
octop-bridge team delete "投研生活组" --yes
```

### 3. Expert Templates & Specialist Instantiation (专家模板库)
```bash
# List all 17+ built-in expert templates
octop-bridge list-experts --json
octop-bridge expert list

# Get details of an expert template
octop-bridge expert get "<expert-template-id>"

# Instantiate a specialist agent from template (e.g. ops-engineer, news-trend, office-automation)
octop-bridge expert create ops-engineer --name "系统运维医生"
octop-bridge expert create news-trend --name "全网热点记者"
```

### 4. Agent Lifecycle & Inspection
```bash
# List active agents
octop-bridge list-agents --json

# Inspect agent runtime status
octop-bridge agent status "<agent-id-or-name>" --json

# Lifecycle control
octop-bridge agent start "<agent-id-or-name>"
octop-bridge agent stop "<agent-id-or-name>"
octop-bridge agent reload "<agent-id-or-name>"
octop-bridge agent delete "<agent-id-or-name>" --yes
```

### 5. Skills & Threads
```bash
# Manage skills on Octop
octop-bridge list-skills --json
octop-bridge skill enable "<skill-name>"
octop-bridge skill disable "<skill-name>"

# List active threads
octop-bridge list-threads --json
```

## Procedure: Creating an Octop Expert Team

When the user asks to create an expert team:
1. **Discover & Confirm Requirements**:
   - Understand what specialists are needed in the team (e.g. travel, stock, ops, news, office automation).
2. **Ensure Member Agents Exist**:
   - Run `octop-bridge list-agents --json` to inspect existing agents.
   - If a required specialist is missing, run `octop-bridge list-experts` to find the corresponding template, and instantiate it with `octop-bridge expert create <template-id> --name "<name>"`.
3. **Assemble the Team**:
   - Run `octop-bridge team create --name "<team_name>" --description "<desc>" --members "<id1,id2,...>"`.
4. **Verify & Interact**:
   - Run `octop-bridge team get "<team_name>"` to confirm members.
   - Run a test probe with `octop-bridge --agent "<team_name>" "<greeting_or_task>" --json` to verify end-to-end responsiveness.

## Pitfalls & Best Practices

- **Team Hosts**: In Octop, an expert team is represented by a host agent (`kind: team`). Chatting with a team addresses the host, which orchestrates members.
- **Roster IDs**: When adding team members via `--members`, comma-separated agent IDs or fuzzy names are accepted; `octop-bridge` automatically resolves names to canonical agent IDs.
- **Latency**: Team dispatch and MCP tool execution take 15–40s; always specify a sufficient timeout (default 180s is safe).

## Verification

- The command exits with code 0.
- JSON output returns `"status": "success"` and non-empty `"response"`.
