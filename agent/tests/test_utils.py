from __future__ import annotations

import unittest

from agent.utils import extract_rank_reference, extract_slots_from_text, format_datetime_text, resolve_date_range


class UtilsTestCase(unittest.TestCase):
    def test_extract_slots(self) -> None:
        slots = extract_slots_from_text("帮我找两张下周六陈奕迅在北京的内场票，预算一共3000元")
        self.assertEqual(slots.city, "北京")
        self.assertEqual(slots.quantity, 2)
        self.assertEqual(slots.budget_total, 3000)
        self.assertEqual(slots.area_preference, "内场")
        self.assertEqual(slots.keyword, "陈奕迅")

    def test_resolve_explicit_date_range(self) -> None:
        start, end = resolve_date_range("2026-04-04", "Asia/Shanghai")
        self.assertEqual(start, "2026-04-04")
        self.assertEqual(end, "2026-04-04")

    def test_extract_rank_reference(self) -> None:
        self.assertEqual(extract_rank_reference("第一场什么时候开票"), 1)
        self.assertEqual(extract_rank_reference("第2个呢"), 2)
        self.assertEqual(extract_rank_reference("最后一场"), -1)

    def test_format_datetime_text(self) -> None:
        self.assertEqual(format_datetime_text("2026-04-04T19:30:00+08:00"), "2026-04-04 19:30")


if __name__ == "__main__":
    unittest.main()
