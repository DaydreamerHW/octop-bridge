"""Data models for Octop Bridge."""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional


@dataclass
class AgentSummary:
    """Metadata summary of an Octop agent/expert."""

    agent_id: str
    name: str
    description: str
    default_model: Optional[str] = None
    enabled: bool = True
    is_shared: bool = False
    icon_name: Optional[str] = None

    @classmethod
    def from_api_dict(cls, data: Dict[str, Any]) -> AgentSummary:
        return cls(
            agent_id=str(data.get("agent_id") or data.get("id") or ""),
            name=str(data.get("name") or "Unnamed"),
            description=str(data.get("description") or ""),
            default_model=data.get("default_model"),
            enabled=bool(data.get("enabled", True)),
            is_shared=bool(data.get("is_shared", False)),
            icon_name=data.get("icon_name"),
        )


@dataclass
class TurnResult:
    """The complete result of an executed turn with an Octop agent."""

    status: str
    agent_id: str
    agent_name: str
    thread_id: Optional[str]
    response: str
    tools_called: List[str] = field(default_factory=list)
    duration_seconds: float = 0.0

    def to_dict(self) -> Dict[str, Any]:
        return {
            "status": self.status,
            "agent_id": self.agent_id,
            "agent_name": self.agent_name,
            "thread_id": self.thread_id,
            "response": self.response,
            "tools_called": self.tools_called,
            "duration_seconds": round(self.duration_seconds, 2),
        }
