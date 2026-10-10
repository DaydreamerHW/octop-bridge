"""Async and sync client for Octop REST and WebSocket APIs."""

from __future__ import annotations

import asyncio
import json
import ssl
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from typing import Any, Callable, Dict, List, Optional

import websockets

from octop_bridge.config import BridgeConfig, load_cached_token, save_cached_token
from octop_bridge.models import AgentSummary, TurnResult


class OctopClient:
    """Client for interacting with self-hosted Octop agent runtimes."""

    def __init__(self, config: BridgeConfig):
        self.config = config
        self._cached_token: Optional[str] = config.token

    def _build_ssl_context(self) -> Optional[ssl.SSLContext]:
        """Build SSL context honoring verify_ssl setting."""
        if not self.config.http_url.startswith("https://") and not self.config.ws_url.startswith("wss://"):
            return None
        ctx = ssl.create_default_context()
        if not self.config.verify_ssl:
            ctx.check_hostname = False
            ctx.verify_mode = ssl.CERT_NONE
        return ctx

    def get_token(self, force_refresh: bool = False) -> str:
        """Retrieve a valid JWT token, using cache when possible."""
        if not force_refresh and self._cached_token:
            return self._cached_token

        # Check file cache
        if not force_refresh:
            cached = load_cached_token(self.config.http_url)
            if cached:
                self._cached_token = cached
                return cached

        # Perform login
        if not self.config.username or not self.config.password:
            raise ValueError("Username and password are required to log in to Octop")

        login_url = f"{self.config.http_url}/api/auth/login"
        payload = {"username": self.config.username, "password": self.config.password}
        data = json.dumps(payload).encode("utf-8")
        req = urllib.request.Request(
            login_url,
            data=data,
            headers={"Content-Type": "application/json"}
        )

        ssl_ctx = self._build_ssl_context()
        try:
            with urllib.request.urlopen(req, timeout=15, context=ssl_ctx) as resp:
                res_data = json.loads(resp.read().decode("utf-8"))
                token = res_data.get("access_token")
                expires_in = res_data.get("expires_in", 86400)
                if not token:
                    raise RuntimeError("Login succeeded but no access_token in response")
                self._cached_token = token
                save_cached_token(self.config.http_url, token, expires_in=expires_in)
                return token
        except urllib.error.HTTPError as e:
            err_msg = e.read().decode("utf-8", errors="replace")
            raise RuntimeError(f"Octop login failed (HTTP {e.code}): {err_msg}") from e
        except Exception as e:
            raise RuntimeError(f"Connection to Octop at {login_url} failed: {e}") from e

    def _api_get(self, path: str) -> Any:
        """Perform an authenticated GET request."""
        token = self.get_token()
        url = f"{self.config.http_url}{path}"
        req = urllib.request.Request(
            url,
            headers={"Authorization": f"Bearer {token}", "Accept": "application/json"}
        )
        ssl_ctx = self._build_ssl_context()
        with urllib.request.urlopen(req, timeout=15, context=ssl_ctx) as resp:
            return json.loads(resp.read().decode("utf-8"))

    def list_agents(self) -> List[AgentSummary]:
        """Fetch list of all agents available on the Octop instance."""
        data = self._api_get("/api/agents")
        agents_raw = data if isinstance(data, list) else data.get("agents", [])
        return [AgentSummary.from_api_dict(a) for a in agents_raw]

    def resolve_agent(self, identifier: Optional[str] = None) -> AgentSummary:
        """Resolve an agent by exact ID, display name, or default fallback."""
        agents = self.list_agents()
        if not agents:
            raise RuntimeError("No agents found on the Octop instance")

        target = (identifier or self.config.default_agent_id or "").strip()
        if not target:
            if len(agents) == 1:
                return agents[0]
            # Try to find default assistant
            for a in agents:
                if "通用" in a.name or "assistant" in a.name.lower():
                    return a
            return agents[0]

        # 1. Exact ID match
        for a in agents:
            if a.agent_id == target:
                return a

        # 2. Exact Name match
        for a in agents:
            if a.name.lower() == target.lower():
                return a

        # 3. Substring Name match
        for a in agents:
            if target.lower() in a.name.lower():
                return a

        # Fallback: if user specified an ID that isn't in list but valid, construct stub
        return AgentSummary(agent_id=target, name=target, description="Specified agent")

    async def chat(
        self,
        query: str,
        agent_id: Optional[str] = None,
        thread_id: Optional[str] = None,
        conversation_mode: str = "craft",
        on_token: Optional[Callable[[str], None]] = None,
        on_tool_start: Optional[Callable[[str], None]] = None,
        on_tool_end: Optional[Callable[[str], None]] = None,
        timeout: Optional[int] = None,
    ) -> TurnResult:
        """Stream a chat conversation turn with the target agent over WebSocket."""
        agent = self.resolve_agent(agent_id)
        token = self.get_token()
        timeout = timeout or self.config.timeout

        ws_url = f"{self.config.ws_url}/api/agents/{agent.agent_id}/chat/ws?token={token}"
        ssl_ctx = self._build_ssl_context()

        turn_payload: Dict[str, Any] = {
            "type": "user_turn",
            "text": query,
            "conversation_mode": conversation_mode,
        }
        if thread_id:
            turn_payload["thread_id"] = thread_id

        full_text: List[str] = []
        tools_called: List[str] = []
        final_thread_id = thread_id
        start_time = time.time()

        async with websockets.connect(ws_url, ssl=ssl_ctx, max_size=32 * 1024 * 1024) as ws:
            await ws.send(json.dumps(turn_payload, ensure_ascii=False))

            while True:
                if time.time() - start_time > timeout:
                    raise TimeoutError(f"Turn with agent {agent.name} timed out after {timeout}s")

                try:
                    raw_msg = await asyncio.wait_for(ws.recv(), timeout=timeout)
                except asyncio.TimeoutError:
                    raise TimeoutError("WebSocket timed out waiting for agent response frame")

                try:
                    frame = json.loads(raw_msg)
                except Exception:
                    continue

                if not isinstance(frame, dict):
                    continue

                f_type = frame.get("type")

                if frame.get("thread_id"):
                    final_thread_id = frame.get("thread_id")

                # Handle streaming tokens
                if f_type == "token":
                    content = frame.get("content") or ""
                    if content:
                        full_text.append(content)
                        if on_token:
                            on_token(content)

                elif f_type in ("chunk", "text_delta"):
                    delta = frame.get("delta") or frame.get("content") or ""
                    if delta:
                        full_text.append(delta)
                        if on_token:
                            on_token(delta)

                # Tool event tracking
                elif f_type in ("tool_start", "tool_call"):
                    name = frame.get("tool") or frame.get("name") or frame.get("tool_name", "unknown")
                    if str(name) not in tools_called:
                        tools_called.append(str(name))
                    if on_tool_start:
                        on_tool_start(str(name))

                elif f_type in ("tool_end", "tool_result"):
                    name = frame.get("tool") or frame.get("name") or ""
                    if on_tool_end:
                        on_tool_end(str(name))

                elif f_type == "state_update":
                    data = frame.get("data")
                    if isinstance(data, dict):
                        for msg in data.get("messages", []):
                            if isinstance(msg, dict):
                                for tc in msg.get("tool_calls", []):
                                    if isinstance(tc, dict):
                                        t_name = tc.get("function", {}).get("name") or tc.get("name")
                                        if t_name and t_name not in tools_called:
                                            tools_called.append(t_name)
                                            if on_tool_start:
                                                on_tool_start(t_name)

                elif f_type in ("done", "turn_done"):
                    break

                elif f_type == "error":
                    err_msg = frame.get("message", "Unknown agent error")
                    raise RuntimeError(f"Agent {agent.name} returned error: {err_msg}")

                elif "message" in frame and isinstance(frame["message"], dict):
                    content = frame["message"].get("content", "")
                    if content and not full_text:
                        full_text.append(content)
                        if on_token:
                            on_token(content)

        duration = time.time() - start_time
        return TurnResult(
            status="success",
            agent_id=agent.agent_id,
            agent_name=agent.name,
            thread_id=final_thread_id,
            response="".join(full_text).strip(),
            tools_called=tools_called,
            duration_seconds=duration,
        )
