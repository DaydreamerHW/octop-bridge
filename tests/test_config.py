"""Unit tests for configuration resolution and models."""

from __future__ import annotations

import os
import unittest
from unittest import mock

from octop_bridge.config import BridgeConfig, resolve_config
from octop_bridge.models import AgentSummary, TurnResult


class TestConfigAndModels(unittest.TestCase):
    def test_bridge_config_urls(self):
        cfg = BridgeConfig(base_url="http://192.168.1.50:8089")
        self.assertEqual(cfg.http_url, "http://192.168.1.50:8089")
        self.assertEqual(cfg.ws_url, "ws://192.168.1.50:8089")

        cfg_https = BridgeConfig(base_url="https://octop.example.com/")
        self.assertEqual(cfg_https.http_url, "https://octop.example.com")
        self.assertEqual(cfg_https.ws_url, "wss://octop.example.com")

    def test_resolve_config_defaults(self):
        with mock.patch.dict(os.environ, {}, clear=True):
            cfg = resolve_config(url="http://test.local:8089")
            self.assertEqual(cfg.http_url, "http://test.local:8089")
            self.assertTrue(cfg.verify_ssl)

    def test_resolve_config_env_overrides(self):
        env = {
            "OCTOP_URL": "https://octop.nas:9000",
            "OCTOP_USERNAME": "admin",
            "OCTOP_PASSWORD": "secret_password",
            "OCTOP_AGENT_ID": "agent_123",
            "OCTOP_INSECURE": "1",
            "OCTOP_TIMEOUT": "60",
        }
        with mock.patch.dict(os.environ, env, clear=True):
            cfg = resolve_config()
            self.assertEqual(cfg.http_url, "https://octop.nas:9000")
            self.assertEqual(cfg.ws_url, "wss://octop.nas:9000")
            self.assertEqual(cfg.username, "admin")
            self.assertEqual(cfg.password, "secret_password")
            self.assertEqual(cfg.default_agent_id, "agent_123")
            self.assertFalse(cfg.verify_ssl)
            self.assertEqual(cfg.timeout, 60)

    def test_agent_summary_model(self):
        raw = {
            "id": "agent_01",
            "name": "Research Assistant",
            "description": "Helps with papers",
            "default_model": "gpt-4o",
            "enabled": True,
        }
        agent = AgentSummary.from_api_dict(raw)
        self.assertEqual(agent.agent_id, "agent_01")
        self.assertEqual(agent.name, "Research Assistant")
        self.assertEqual(agent.default_model, "gpt-4o")
        self.assertTrue(agent.enabled)

    def test_turn_result_serialization(self):
        res = TurnResult(
            status="success",
            agent_id="PQG34P",
            agent_name="小通",
            thread_id="thr_123",
            response="Done",
            tools_called=["search_flights"],
            duration_seconds=1.234,
        )
        d = res.to_dict()
        self.assertEqual(d["status"], "success")
        self.assertEqual(d["agent_name"], "小通")
        self.assertEqual(d["duration_seconds"], 1.23)
        self.assertEqual(d["tools_called"], ["search_flights"])


if __name__ == "__main__":
    unittest.main()
