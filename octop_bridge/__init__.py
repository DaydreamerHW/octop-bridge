"""Octop Bridge: Unified client & AgentSkill bridge for Octop AI agents."""

from __future__ import annotations

from octop_bridge.client import OctopClient
from octop_bridge.config import BridgeConfig, resolve_config
from octop_bridge.models import AgentSummary, TurnResult

__version__ = "0.2.0"

__all__ = [
    "OctopClient",
    "BridgeConfig",
    "resolve_config",
    "AgentSummary",
    "TurnResult",
    "__version__",
]
