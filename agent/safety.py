from __future__ import annotations

from dataclasses import dataclass


TICKET_KEYWORDS = {
    "票",
    "门票",
    "购票",
    "开票",
    "售票",
    "演唱会",
    "演出",
    "场馆",
    "场次",
    "实名",
    "退票",
    "入场",
    "候补",
    "内场",
    "看台",
    "连座",
    "票档",
    "音乐节",
    "座位",
}

GREETING_KEYWORDS = {"你好", "您好", "hi", "hello", "在吗", "嘿"}
FOLLOW_UP_KEYWORDS = {"那", "那个", "这场", "杭州的呢", "北京的呢", "还有吗", "还有票吗"}
BLOCKLIST_KEYWORDS = {
    "写代码",
    "帮我编程",
    "python",
    "java",
    "股票",
    "基金",
    "法律意见",
    "医疗建议",
    "算命",
    "情色",
    "违规",
}


@dataclass(slots=True)
class SafetyDecision:
    allowed: bool
    reason: str
    response: str | None = None


class TicketingSafetyGuard:
    def inspect(self, text: str, *, has_context: bool = False) -> SafetyDecision:
        normalized = text.strip().lower()
        if not normalized:
            return SafetyDecision(False, "empty", "请告诉我想查询的演出、城市、时间或票务规则。")

        if any(keyword in normalized for keyword in GREETING_KEYWORDS):
            return SafetyDecision(True, "greeting")

        if has_context and any(keyword in text for keyword in FOLLOW_UP_KEYWORDS):
            return SafetyDecision(True, "follow_up")

        has_ticket_keyword = any(keyword in text for keyword in TICKET_KEYWORDS)
        has_blocklist_keyword = any(keyword in normalized for keyword in BLOCKLIST_KEYWORDS)

        if has_blocklist_keyword and not has_ticket_keyword:
            return SafetyDecision(
                False,
                "out_of_scope",
                "我现在只负责票务相关的查询、规则解释和购票路径引导，其他话题就先不展开了。",
            )

        if has_ticket_keyword:
            return SafetyDecision(True, "ticket_related")

        if len(text) <= 12 and has_context:
            return SafetyDecision(True, "short_follow_up")

        return SafetyDecision(
            False,
            "not_ticket_related",
            "我现在专注于票务问答，可以帮你查演出、解释实名/退票规则，或者根据预算推荐合适票档。",
        )
