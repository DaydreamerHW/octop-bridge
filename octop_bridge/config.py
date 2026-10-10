"""Configuration and token cache management for Octop Bridge."""

from __future__ import annotations

import json
import os
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any, Dict, Optional
from urllib.parse import urlparse

CONFIG_DIR = Path.home() / ".config" / "octop-bridge"
CONFIG_FILE = CONFIG_DIR / "config.json"
CACHE_DIR = Path.home() / ".cache" / "octop-bridge"
TOKEN_CACHE_FILE = CACHE_DIR / "token_cache.json"


@dataclass
class BridgeConfig:
    """Consolidated configuration for connecting to an Octop instance."""

    base_url: str = "http://127.0.0.1:8089"
    username: Optional[str] = None
    password: Optional[str] = None
    token: Optional[str] = None
    default_agent_id: Optional[str] = None
    verify_ssl: bool = True
    timeout: int = 180

    @property
    def http_url(self) -> str:
        """Normalized HTTP(S) base URL without trailing slash."""
        url = self.base_url.rstrip("/")
        if not url.startswith(("http://", "https://")):
            url = f"http://{url}"
        return url

    @property
    def ws_url(self) -> str:
        """WebSocket base URL derived from HTTP URL."""
        parsed = urlparse(self.http_url)
        ws_scheme = "wss" if parsed.scheme == "https" else "ws"
        return f"{ws_scheme}://{parsed.netloc}"


def load_file_config() -> Dict[str, Any]:
    """Load configuration from ~/.config/octop-bridge/config.json if it exists."""
    if not CONFIG_FILE.is_file():
        return {}
    try:
        with open(CONFIG_FILE, "r", encoding="utf-8") as f:
            return json.load(f)
    except Exception:
        return {}


def save_file_config(cfg: Dict[str, Any]) -> None:
    """Save persistent configuration to ~/.config/octop-bridge/config.json."""
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    with open(CONFIG_FILE, "w", encoding="utf-8") as f:
        json.dump(cfg, f, indent=2, ensure_ascii=False)


def load_cached_token(base_url: str) -> Optional[str]:
    """Retrieve unexpired JWT token for the given base_url from cache."""
    if not TOKEN_CACHE_FILE.is_file():
        return None
    try:
        with open(TOKEN_CACHE_FILE, "r", encoding="utf-8") as f:
            cache = json.load(f)
            entry = cache.get(base_url)
            if not entry:
                return None
            token = entry.get("token")
            expires_at = entry.get("expires_at", 0)
            # Give 60 seconds buffer before expiration
            if token and expires_at > (os.times().system + 60 if hasattr(os.times(), "system") else 0):
                import time
                if expires_at > (time.time() + 60):
                    return token
    except Exception:
        return None
    return None


def save_cached_token(base_url: str, token: str, expires_in: int = 86400) -> None:
    """Cache the retrieved JWT token with expiration timestamp."""
    import time
    CACHE_DIR.mkdir(parents=True, exist_ok=True)
    cache: Dict[str, Any] = {}
    if TOKEN_CACHE_FILE.is_file():
        try:
            with open(TOKEN_CACHE_FILE, "r", encoding="utf-8") as f:
                cache = json.load(f)
        except Exception:
            cache = {}

    cache[base_url] = {
        "token": token,
        "expires_at": time.time() + expires_in,
        "cached_at": time.time(),
    }
    try:
        with open(TOKEN_CACHE_FILE, "w", encoding="utf-8") as f:
            json.dump(cache, f, indent=2)
    except Exception:
        pass


def resolve_config(
    url: Optional[str] = None,
    username: Optional[str] = None,
    password: Optional[str] = None,
    token: Optional[str] = None,
    agent_id: Optional[str] = None,
    insecure: bool = False,
    timeout: Optional[int] = None,
) -> BridgeConfig:
    """Resolve configuration with hierarchy: Explicit Flags > Env Vars > Config File > Defaults."""
    file_cfg = load_file_config()

    # 1. Base URL
    resolved_url = (
        url
        or os.environ.get("OCTOP_URL")
        or (
            f"http://{os.environ.get('OCTOP_HOST', '192.168.5.192')}:{os.environ.get('OCTOP_PORT', '8089')}"
            if os.environ.get("OCTOP_HOST")
            else None
        )
        or file_cfg.get("base_url")
        or "http://192.168.5.192:8089"  # Default LAN host for this deployment
    )

    # 2. Token
    resolved_token = token or os.environ.get("OCTOP_TOKEN") or file_cfg.get("token")

    # 3. Credentials
    resolved_username = username or os.environ.get("OCTOP_USERNAME") or file_cfg.get("username") or "xqx"
    resolved_password = password or os.environ.get("OCTOP_PASSWORD") or file_cfg.get("password") or "Xqxq@1102@2026"

    # 4. Default Agent ID
    resolved_agent = agent_id or os.environ.get("OCTOP_AGENT_ID") or file_cfg.get("default_agent_id") or "PQG34P"

    # 5. SSL & Timeout
    verify_ssl = not (
        insecure
        or os.environ.get("OCTOP_INSECURE", "").lower() in ("1", "true", "yes")
        or file_cfg.get("insecure", False)
    )
    resolved_timeout = timeout or int(os.environ.get("OCTOP_TIMEOUT", file_cfg.get("timeout", 180)))

    return BridgeConfig(
        base_url=resolved_url,
        username=resolved_username,
        password=resolved_password,
        token=resolved_token,
        default_agent_id=resolved_agent,
        verify_ssl=verify_ssl,
        timeout=resolved_timeout,
    )
