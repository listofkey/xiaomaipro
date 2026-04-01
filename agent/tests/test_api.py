from __future__ import annotations

import unittest

from fastapi.testclient import TestClient

from agent.config import Settings
from agent.main import create_app


class ChatApiTestCase(unittest.TestCase):
    def setUp(self) -> None:
        settings = Settings(
            gateway_mode="mock",
            bootstrap_knowledge_on_startup=False,
            openai_api_key=None,
            embedding_api_key=None,
            app_reference_date="2026-03-27",
        )
        self.client = TestClient(create_app(settings))
        self.client.__enter__()

    def tearDown(self) -> None:
        self.client.__exit__(None, None, None)

    def test_search_chat_returns_cards(self) -> None:
        response = self.client.post(
            "/api/v1/chat",
            json={
                "session_id": "ticket-session",
                "message": "帮我找两张下周六陈奕迅在北京的内场票，预算一共3000元",
            },
        )
        self.assertEqual(response.status_code, 200)
        payload = response.json()
        self.assertTrue(payload["fallback_mode"])
        self.assertGreaterEqual(len(payload["cards"]), 1)
        self.assertIn("陈奕迅", payload["answer"])
        self.assertIn("内场票已售罄", payload["answer"])

    def test_follow_up_keeps_context(self) -> None:
        self.client.post(
            "/api/v1/chat",
            json={
                "session_id": "follow-up-session",
                "message": "周杰伦在上海的票",
            },
        )
        response = self.client.post(
            "/api/v1/chat",
            json={
                "session_id": "follow-up-session",
                "message": "那杭州的呢？",
            },
        )
        self.assertEqual(response.status_code, 200)
        payload = response.json()
        self.assertEqual(payload["slots"]["city"], "杭州")

    def test_follow_up_can_query_event_sale_time(self) -> None:
        self.client.post(
            "/api/v1/chat",
            json={
                "session_id": "detail-follow-up-session",
                "message": "帮我找两张下周六陈奕迅在北京的内场票，预算一共3000元",
            },
        )
        response = self.client.post(
            "/api/v1/chat",
            json={
                "session_id": "detail-follow-up-session",
                "message": "第一场什么时候开票？",
            },
        )
        self.assertEqual(response.status_code, 200)
        payload = response.json()
        self.assertIn("2026-03-28 12:00", payload["answer"])
        self.assertEqual(payload["cards"][0]["event_id"], "1001")
        self.assertEqual(payload["tools"][0]["name"], "query_event_detail")

    def test_follow_up_can_query_real_name_requirement(self) -> None:
        self.client.post(
            "/api/v1/chat",
            json={
                "session_id": "real-name-follow-up-session",
                "message": "周杰伦在上海的票",
            },
        )
        response = self.client.post(
            "/api/v1/chat",
            json={
                "session_id": "real-name-follow-up-session",
                "message": "需要实名吗？",
            },
        )
        self.assertEqual(response.status_code, 200)
        payload = response.json()
        self.assertIn("实名制", payload["answer"])
        self.assertEqual(payload["cards"][0]["event_id"], "1002")

    def test_stream_endpoint_emits_start_event(self) -> None:
        with self.client.stream(
            "POST",
            "/api/v1/chat/stream",
            json={
                "session_id": "stream-session",
                "message": "周杰伦在上海的票",
            },
        ) as response:
            content = "".join(response.iter_text())
        self.assertEqual(response.status_code, 200)
        self.assertIn("event: start", content)
        self.assertIn("event: done", content)


if __name__ == "__main__":
    unittest.main()
