"""Command-line interface for Octop Bridge."""

from __future__ import annotations

import argparse
import asyncio
import json
import sys
from typing import Optional

from octop_bridge.client import OctopClient
from octop_bridge.config import resolve_config, save_file_config


def cmd_list_agents(args: argparse.Namespace) -> int:
    """List all agents available on the configured Octop instance."""
    cfg = resolve_config(
        url=args.url,
        username=args.username,
        password=args.password,
        token=args.token,
        insecure=args.insecure,
    )
    client = OctopClient(cfg)
    try:
        agents = client.list_agents()
    except Exception as e:
        sys.stderr.write(f"Error fetching agents: {e}\n")
        return 1

    if args.json:
        print(json.dumps([a.__dict__ for a in agents], ensure_ascii=False, indent=2))
        return 0

    print(f"\nConnected to Octop: {cfg.http_url}")
    print(f"Found {len(agents)} agent(s):\n")
    print(f"{'ID':<10} | {'NAME':<20} | {'MODEL':<28} | {'STATUS':<8}")
    print("-" * 75)
    for a in agents:
        status = "Active" if a.enabled else "Disabled"
        model = a.default_model or "Default"
        print(f"{a.agent_id:<10} | {a.name:<20} | {model:<28} | {status:<8}")
    print()
    return 0


def cmd_list_threads(args: argparse.Namespace) -> int:
    """List historical conversation threads for the agent."""
    cfg = resolve_config(
        url=args.url,
        username=args.username,
        password=args.password,
        token=args.token,
        agent_id=args.agent,
        insecure=args.insecure,
    )
    client = OctopClient(cfg)
    try:
        agent = client.resolve_agent(args.agent)
        threads = client.list_threads(agent.agent_id)
    except Exception as e:
        sys.stderr.write(f"Error fetching threads: {e}\n")
        return 1

    if args.json:
        print(json.dumps(threads, ensure_ascii=False, indent=2))
        return 0

    print(f"\nThreads for Agent: {agent.name} ({agent.agent_id})")
    print(f"Total: {len(threads)} thread(s):\n")
    print(f"{'THREAD ID':<32} | {'PREVIEW'}")
    print("-" * 75)
    for t in threads:
        tid = t.get("thread_id", "")
        preview = (t.get("preview") or t.get("title") or "").replace("\n", " ")[:38]
        print(f"{tid:<32} | {preview}")
    print()
    return 0


def cmd_setup(args: argparse.Namespace) -> int:
    """Interactive wizard to configure Octop Bridge connection settings."""
    print("=== Octop Bridge Setup Wizard ===\n")
    url = input("Octop Server URL [http://192.168.5.192:8089]: ").strip() or "http://192.168.5.192:8089"
    username = input("Username [xqx]: ").strip() or "xqx"
    password = input("Password [leave empty to keep unchanged]: ").strip() or "Xqxq@1102@2026"
    agent_id = input("Default Agent ID or Name [PQG34P]: ").strip() or "PQG34P"

    cfg_dict = {
        "base_url": url,
        "username": username,
        "password": password,
        "default_agent_id": agent_id,
    }
    save_file_config(cfg_dict)
    print("\n✓ Configuration saved to ~/.config/octop-bridge/config.json")
    return 0


def cmd_chat(args: argparse.Namespace) -> int:
    """Send a query to the specified Octop agent and stream response."""
    cfg = resolve_config(
        url=args.url,
        username=args.username,
        password=args.password,
        token=args.token,
        agent_id=args.agent,
        insecure=args.insecure,
        timeout=args.timeout,
    )
    client = OctopClient(cfg)

    prompt = args.query
    if not prompt:
        sys.stderr.write("Error: Query prompt is required.\n")
        return 1

    stream = not args.no_stream and not args.json

    def on_token(delta: str) -> None:
        if stream:
            sys.stdout.write(delta)
            sys.stdout.flush()

    def on_tool_start(tool_name: str) -> None:
        if stream:
            sys.stdout.write(f"\n[小通正在调用工具: {tool_name}]...\n")
            sys.stdout.flush()

    def on_tool_end(tool_name: str) -> None:
        if stream:
            sys.stdout.write(f"[工具调用完成: {tool_name}]\n")
            sys.stdout.flush()

    try:
        result = asyncio.run(
            client.chat(
                query=prompt,
                agent_id=args.agent,
                thread_id=args.thread_id,
                new_thread=args.new_thread,
                conversation_mode=args.mode,
                on_token=on_token,
                on_tool_start=on_tool_start,
                on_tool_end=on_tool_end,
                timeout=args.timeout,
            )
        )
        if stream:
            sys.stdout.write("\n")
            sys.stdout.flush()

        if args.json:
            print(json.dumps(result.to_dict(), ensure_ascii=False, indent=2))
        elif not stream:
            print(result.response)
        return 0
    except Exception as e:
        sys.stderr.write(f"\nError: {e}\n")
        return 1


def main(argv: Optional[list[str]] = None) -> int:
    """Main CLI dispatch."""
    parser = argparse.ArgumentParser(
        prog="octop-bridge",
        description="Octop Bridge: Unified CLI client to chat and delegate tasks to Octop AI agents",
    )
    # Global options
    parser.add_argument("--url", default=None, help="Octop server base URL (e.g. http://192.168.5.192:8089)")
    parser.add_argument("--username", "-u", default=None, help="Octop login username")
    parser.add_argument("--password", "-p", default=None, help="Octop login password")
    parser.add_argument("--token", default=None, help="Direct JWT Access Token (skips login)")
    parser.add_argument("--insecure", "-k", action="store_true", help="Allow insecure TLS/SSL connections")
    parser.add_argument("--json", action="store_true", help="Output result as JSON")

    subparsers = parser.add_subparsers(dest="command", help="Sub-commands")

    # Subcommand: list-agents
    sub_list = subparsers.add_parser("list-agents", aliases=["ls"], help="List available agents on Octop")
    sub_list.set_defaults(func=cmd_list_agents)

    # Subcommand: list-threads
    sub_threads = subparsers.add_parser("list-threads", aliases=["threads"], help="List conversation threads for an agent")
    sub_threads.add_argument("--agent", "-a", default=None, help="Agent ID or Display Name")
    sub_threads.set_defaults(func=cmd_list_threads)

    # Subcommand: setup
    sub_setup = subparsers.add_parser("setup", help="Configure server connection interactively")
    sub_setup.set_defaults(func=cmd_setup)

    # Subcommand: chat
    sub_chat = subparsers.add_parser("chat", help="Send a message to an Octop agent")
    sub_chat.add_argument("query", help="Prompt or task to send")
    sub_chat.add_argument("--agent", "-a", default=None, help="Agent ID or Display Name (fuzzy matched)")
    sub_chat.add_argument("--thread-id", "-t", default=None, help="Conversation thread ID to continue")
    sub_chat.add_argument("--new", "-n", dest="new_thread", action="store_true", help="Start a new isolated conversation thread")
    sub_chat.add_argument("--mode", "-m", choices=["craft", "ask", "plan"], default="craft", help="Turn mode")
    sub_chat.add_argument("--timeout", type=int, default=None, help="Turn timeout in seconds")
    sub_chat.add_argument("--no-stream", action="store_true", help="Disable live streaming output")
    sub_chat.set_defaults(func=cmd_chat)

    # Shorthand: direct query without explicit "chat" subcommand
    args_list = argv if argv is not None else sys.argv[1:]
    if args_list and args_list[0] not in ("list-agents", "ls", "list-threads", "threads", "setup", "chat", "-h", "--help"):
        # Wrap into chat subcommand
        args_list = ["chat"] + args_list

    args = parser.parse_args(args_list)

    if not hasattr(args, "func"):
        parser.print_help()
        return 0

    return args.func(args)


if __name__ == "__main__":
    sys.exit(main())
