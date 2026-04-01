from __future__ import annotations

import json
import re
from datetime import date, datetime, timedelta
from typing import Any
from zoneinfo import ZoneInfo, ZoneInfoNotFoundError

from agent.schemas import SlotExtraction


KNOWN_CITIES = [
    "北京",
    "上海",
    "广州",
    "深圳",
    "杭州",
    "成都",
    "重庆",
    "南京",
    "苏州",
    "武汉",
    "长沙",
    "西安",
    "天津",
    "青岛",
    "郑州",
]

AREA_KEYWORDS = ["内场", "看台", "前排", "后排", "VIP", "vip", "包厢", "连座"]
STOPWORDS = {
    "帮我",
    "帮忙",
    "看看",
    "查查",
    "找找",
    "两张",
    "一张",
    "三张",
    "四张",
    "五张",
    "预算",
    "总共",
    "一共",
    "下周",
    "这周",
    "本周",
    "周六",
    "周日",
    "周末",
    "演唱会",
    "音乐节",
    "门票",
    "票务",
    "购票",
    "场次",
    "还有",
    "有没有",
}

CN_NUM_MAP = {
    "零": 0,
    "一": 1,
    "二": 2,
    "两": 2,
    "三": 3,
    "四": 4,
    "五": 5,
    "六": 6,
    "七": 7,
    "八": 8,
    "九": 9,
    "十": 10,
}

WEEKDAY_MAP = {
    "一": 0,
    "二": 1,
    "三": 2,
    "四": 3,
    "五": 4,
    "六": 5,
    "日": 6,
    "天": 6,
}

RANK_PATTERNS: list[tuple[re.Pattern[str], int | str]] = [
    (re.compile(r"第?\s*1\s*(?:个|场)?"), 1),
    (re.compile(r"第?\s*2\s*(?:个|场)?"), 2),
    (re.compile(r"第?\s*3\s*(?:个|场)?"), 3),
    (re.compile(r"第?\s*4\s*(?:个|场)?"), 4),
    (re.compile(r"第?\s*5\s*(?:个|场)?"), 5),
    (re.compile(r"第一(?:个|场)?"), 1),
    (re.compile(r"第二(?:个|场)?"), 2),
    (re.compile(r"第三(?:个|场)?"), 3),
    (re.compile(r"第四(?:个|场)?"), 4),
    (re.compile(r"第五(?:个|场)?"), 5),
    (re.compile(r"最后一(?:个|场)"), "last"),
]


def json_dumps(data: Any) -> str:
    return json.dumps(data, ensure_ascii=False, separators=(",", ":"))


def now_in_timezone(timezone_name: str) -> datetime:
    try:
        return datetime.now(ZoneInfo(timezone_name))
    except ZoneInfoNotFoundError:
        if timezone_name == "Asia/Shanghai":
            return datetime.utcnow() + timedelta(hours=8)
        return datetime.utcnow()


def truncate(text: str, limit: int = 140) -> str:
    clean = " ".join(text.split())
    if len(clean) <= limit:
        return clean
    return clean[: limit - 3] + "..."


def format_datetime_text(value: str | None) -> str:
    if not value:
        return "待定"
    text = value.strip()
    for fmt in ("%Y-%m-%dT%H:%M:%S%z", "%Y-%m-%dT%H:%M:%S", "%Y-%m-%d %H:%M:%S"):
        try:
            return datetime.strptime(text, fmt).strftime("%Y-%m-%d %H:%M")
        except ValueError:
            continue
    if "T" in text:
        return text.replace("T", " ")[:16]
    return text[:16]


def chinese_number_to_int(text: str) -> int | None:
    if not text:
        return None
    if text.isdigit():
        return int(text)
    if text == "十":
        return 10
    if len(text) == 2 and text[0] == "十":
        return 10 + CN_NUM_MAP.get(text[1], 0)
    if len(text) == 2 and text[1] == "十":
        return CN_NUM_MAP.get(text[0], 0) * 10
    if len(text) == 3 and text[1] == "十":
        return CN_NUM_MAP.get(text[0], 0) * 10 + CN_NUM_MAP.get(text[2], 0)
    return CN_NUM_MAP.get(text)


def extract_rank_reference(text: str) -> int | None:
    normalized = text.strip()
    for pattern, value in RANK_PATTERNS:
        if pattern.search(normalized):
            if value == "last":
                return -1
            return int(value)
    return None


def extract_slots_from_text(text: str) -> SlotExtraction:
    lowered = text.lower()
    quantity = None
    quantity_match = re.search(r"([一二两三四五六七八九十\d]+)\s*张", text)
    if quantity_match:
        quantity = chinese_number_to_int(quantity_match.group(1))

    budget_total = None
    total_match = re.search(r"(?:预算|一共|总共|总价)\D{0,3}(\d{2,6})\s*元?", text)
    if total_match:
        budget_total = float(total_match.group(1))

    budget_per_ticket = None
    per_match = re.search(r"(?:每张|单张|票价)\D{0,3}(\d{2,6})\s*元?", text)
    if per_match:
        budget_per_ticket = float(per_match.group(1))

    city = next((item for item in KNOWN_CITIES if item in text), None)
    area = next((item for item in AREA_KEYWORDS if item.lower() in lowered), None)
    date_hint = extract_date_hint(text)
    keyword = extract_keyword(text, city=city, area=area, date_hint=date_hint)

    return SlotExtraction(
        city=city,
        quantity=quantity,
        budget_total=budget_total,
        budget_per_ticket=budget_per_ticket,
        area_preference=area,
        date_hint=date_hint,
        keyword=keyword,
        artist=keyword,
    )


def extract_date_hint(text: str) -> str | None:
    direct_patterns = [
        r"\d{4}-\d{2}-\d{2}",
        r"\d{4}/\d{2}/\d{2}",
        r"(?:今天|明天|后天|本周末|这周末|下周末|本周[一二三四五六日天]|这周[一二三四五六日天]|下周[一二三四五六日天]|周末|周[一二三四五六日天]|本月|下个月|下月)",
    ]
    for pattern in direct_patterns:
        match = re.search(pattern, text)
        if match:
            return match.group(0)
    return None


def extract_keyword(
    text: str,
    *,
    city: str | None = None,
    area: str | None = None,
    date_hint: str | None = None,
) -> str | None:
    stripped = text
    for token in [city, area, date_hint, "门票", "票", "演出", "演唱会", "音乐节", "周末", "下周", "这周", "本周"]:
        if token:
            stripped = stripped.replace(token, " ")
    for token in sorted(STOPWORDS, key=len, reverse=True):
        stripped = stripped.replace(token, " ")
    stripped = re.sub(r"\d+", " ", stripped)
    stripped = re.sub(r"[的呢吗呀啊在去看找帮让给个张场次预算总共一共]", " ", stripped)
    stripped = re.sub(r"[^\u4e00-\u9fa5A-Za-z0-9·]+", " ", stripped)
    candidates = [chunk.strip() for chunk in stripped.split() if len(chunk.strip()) >= 2]
    candidates = [item for item in candidates if item not in STOPWORDS]
    if not candidates:
        return None
    return max(candidates, key=len)


def resolve_date_range(
    date_hint: str | None,
    timezone_name: str,
    reference_date: str | None = None,
) -> tuple[str | None, str | None]:
    if not date_hint:
        return None, None

    today = now_in_timezone(timezone_name).date()
    if reference_date:
        try:
            today = datetime.strptime(reference_date, "%Y-%m-%d").date()
        except ValueError:
            pass
    hint = date_hint.strip()

    for fmt in ("%Y-%m-%d", "%Y/%m/%d"):
        try:
            parsed = datetime.strptime(hint, fmt).date()
            return parsed.isoformat(), parsed.isoformat()
        except ValueError:
            continue

    if hint == "今天":
        return today.isoformat(), today.isoformat()
    if hint == "明天":
        target = today + timedelta(days=1)
        return target.isoformat(), target.isoformat()
    if hint == "后天":
        target = today + timedelta(days=2)
        return target.isoformat(), target.isoformat()
    if hint in {"本月", "这个月"}:
        start = today.replace(day=1)
        end = _end_of_month(start)
        return start.isoformat(), end.isoformat()
    if hint in {"下个月", "下月"}:
        month_start = (today.replace(day=28) + timedelta(days=4)).replace(day=1)
        return month_start.isoformat(), _end_of_month(month_start).isoformat()
    if hint in {"周末", "本周末", "这周末"}:
        saturday = today + timedelta(days=(5 - today.weekday()) % 7)
        sunday = saturday + timedelta(days=1)
        return saturday.isoformat(), sunday.isoformat()
    if hint == "下周末":
        week_start = today - timedelta(days=today.weekday()) + timedelta(days=7)
        saturday = week_start + timedelta(days=5)
        sunday = saturday + timedelta(days=1)
        return saturday.isoformat(), sunday.isoformat()

    week_match = re.fullmatch(r"(本|这|下)?周([一二三四五六日天])", hint)
    if week_match:
        prefix, weekday_text = week_match.groups()
        weekday_index = WEEKDAY_MAP[weekday_text]
        week_start = today - timedelta(days=today.weekday())
        if prefix == "下":
            week_start = week_start + timedelta(days=7)
        target = week_start + timedelta(days=weekday_index)
        return target.isoformat(), target.isoformat()

    return None, None


def _end_of_month(value: date) -> date:
    next_month = (value.replace(day=28) + timedelta(days=4)).replace(day=1)
    return next_month - timedelta(days=1)


def sse_event(event: str, data: dict[str, Any]) -> bytes:
    return f"event: {event}\ndata: {json_dumps(data)}\n\n".encode("utf-8")
