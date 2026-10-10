"""Unit tests for OctopClient agent resolution."""

from __future__ import annotations

import unittest
from unittest import mock

from octop_bridge.client import OctopClient
from octop_bridge.config import BridgeConfig
from octop_bridge.models import AgentSummary


class TestOctopClient(unittest.TestCase):
    def test_resolve_agent_matching(self):
        cfg = BridgeConfig(base_url="http://test.local:8089")
        client = OctopClient(cfg)

        mock_agents = [
            AgentSummary(agent_id="PQG34P", name="小通 · 通用助手", description="Default assistant"),
            AgentSummary(agent_id="COD123", name="Code Specialist", description="Writes python"),
        ]

        with mock.patch.object(client, "list_agents", return_value=mock_agents):
            # Match by ID
            a1 = client.resolve_agent("PQG34P")
            self.assertEqual(a1.agent_id, "PQG34P")

            # Match by exact name
            a2 = client.resolve_agent("Code Specialist")
            self.assertEqual(a2.agent_id, "COD123")

            # Match by partial/fuzzy name
            a3 = client.resolve_agent("小通")
            self.assertEqual(a3.agent_id, "PQG34P")

            # Default fallback when target is None
            a4 = client.resolve_agent(None)
            self.assertEqual(a4.agent_id, "PQG34P")


if __name__ == "__main__":
    unittest.main()
