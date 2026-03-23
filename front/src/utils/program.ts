import type {
  ProgramCategoryInfo,
  ProgramEventBrief,
  ProgramEventDetail,
  ProgramTicketTierInfo,
} from '@/api/program'

export type ProgramPageKind = 'all' | 'concert' | 'festival' | 'drama' | 'sports' | 'exhibition'
export type TimeFilter = 'all' | 'today' | 'tomorrow' | 'weekend' | 'next30'

type TicketState = {
  disabled: boolean
  label: string
  tagType: 'success' | 'warning' | 'info' | 'danger'
}

const pageTitleMap: Record<ProgramPageKind, string> = {
  all: '精选活动',
  concert: '演唱会',
  festival: '音乐节',
  drama: '话剧演出',
  sports: '体育赛事',
  exhibition: '展览休闲',
}

const fallbackCategoryIdMap: Record<Exclude<ProgramPageKind, 'all'>, number> = {
  concert: 1,
  drama: 2,
  festival: 3,
  sports: 4,
  exhibition: 5,
}

const categoryKeywords: Record<Exclude<ProgramPageKind, 'all'>, string[]> = {
  concert: ['演唱会', 'concert', 'live'],
  drama: ['话剧', '戏剧', '舞剧', '歌剧', 'drama', 'musical', 'theatre'],
  festival: ['音乐节', 'festival'],
  sports: ['体育', 'sports', 'match'],
  exhibition: ['展览', '展会', '展演', 'exhibition', 'exhibit'],
}

function normalizeText(value?: string) {
  return (value || '').toLowerCase().replace(/\s+/g, '')
}

function parseDateTime(value?: string) {
  if (!value) {
    return null
  }

  const parsed = new Date(value.replace(' ', 'T'))
  if (Number.isNaN(parsed.getTime())) {
    return null
  }

  return parsed
}

function formatByOptions(value?: string, options?: Intl.DateTimeFormatOptions) {
  const parsed = parseDateTime(value)
  if (!parsed) {
    return ''
  }

  return new Intl.DateTimeFormat('zh-CN', options).format(parsed)
}

function addDays(base: Date, days: number) {
  const next = new Date(base)
  next.setDate(next.getDate() + days)
  return next
}

function toDateString(value: Date) {
  return value.toISOString().slice(0, 10)
}

function createPosterFallback(title: string) {
  const safeTitle = title || 'Ticket Pro'
  const svg = `
    <svg xmlns="http://www.w3.org/2000/svg" width="640" height="860" viewBox="0 0 640 860">
      <defs>
        <linearGradient id="g" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#0f766e"/>
          <stop offset="100%" stop-color="#1d4ed8"/>
        </linearGradient>
      </defs>
      <rect width="640" height="860" fill="url(#g)"/>
      <circle cx="520" cy="150" r="120" fill="rgba(255,255,255,0.14)"/>
      <circle cx="120" cy="700" r="180" fill="rgba(255,255,255,0.12)"/>
      <text x="56" y="130" fill="#ffffff" font-size="28" font-family="sans-serif" opacity="0.85">Ticket Pro</text>
      <foreignObject x="56" y="220" width="528" height="420">
        <div xmlns="http://www.w3.org/1999/xhtml" style="color:#fff;font-size:46px;line-height:1.35;font-family:sans-serif;font-weight:700;">
          ${safeTitle}
        </div>
      </foreignObject>
    </svg>
  `

  return `data:image/svg+xml;charset=UTF-8,${encodeURIComponent(svg)}`
}

export function getPageTitle(kind: ProgramPageKind) {
  return pageTitleMap[kind]
}

export function resolveCategoryIdForPage(categories: ProgramCategoryInfo[], kind: ProgramPageKind) {
  if (kind === 'all') {
    return 0
  }

  const matched = categories.find((item) => {
    const normalized = normalizeText(item.name)
    return categoryKeywords[kind].some((keyword) => normalized.includes(normalizeText(keyword)))
  })

  return matched?.id ?? fallbackCategoryIdMap[kind]
}

export function buildDateRangeByFilter(filter: TimeFilter) {
  if (filter === 'all') {
    return {}
  }

  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())

  if (filter === 'today') {
    const date = toDateString(today)
    return { startDate: date, endDate: date }
  }

  if (filter === 'tomorrow') {
    const tomorrow = addDays(today, 1)
    const date = toDateString(tomorrow)
    return { startDate: date, endDate: date }
  }

  if (filter === 'weekend') {
    const day = today.getDay()
    const saturdayOffset = day === 6 ? 0 : (6 - day + 7) % 7
    const saturday = addDays(today, saturdayOffset)
    const sunday = addDays(saturday, 1)
    return {
      startDate: toDateString(saturday),
      endDate: toDateString(sunday),
    }
  }

  const end = addDays(today, 30)
  return {
    startDate: toDateString(today),
    endDate: toDateString(end),
  }
}

export function resolvePosterUrl(posterUrl?: string, title?: string) {
  if (posterUrl) {
    return posterUrl
  }

  return createPosterFallback(title || '')
}

export function formatCurrency(value?: number) {
  if (typeof value !== 'number' || Number.isNaN(value)) {
    return '0'
  }

  return new Intl.NumberFormat('zh-CN', {
    minimumFractionDigits: Number.isInteger(value) ? 0 : 2,
    maximumFractionDigits: 2,
  }).format(value)
}

export function formatDateTimeLabel(value?: string) {
  return formatByOptions(value, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatEventCardTime(start?: string, end?: string) {
  const startDate = parseDateTime(start)
  const endDate = parseDateTime(end)
  if (!startDate || !endDate) {
    return '时间待定'
  }

  const sameDay =
    startDate.getFullYear() === endDate.getFullYear() &&
    startDate.getMonth() === endDate.getMonth() &&
    startDate.getDate() === endDate.getDate()

  if (sameDay) {
    return `${formatByOptions(start, {
      month: '2-digit',
      day: '2-digit',
      weekday: 'short',
      hour: '2-digit',
      minute: '2-digit',
    })} 开演`
  }

  return `${formatByOptions(start, {
    month: '2-digit',
    day: '2-digit',
  })} - ${formatByOptions(end, {
    month: '2-digit',
    day: '2-digit',
  })}`
}

export function formatEventDetailTime(start?: string, end?: string) {
  const startText = formatDateTimeLabel(start)
  const endText = formatDateTimeLabel(end)
  if (!startText && !endText) {
    return '时间待定'
  }
  if (!endText || startText === endText) {
    return startText || endText
  }
  return `${startText} - ${endText}`
}

export function getEventStatusLabel(status?: number) {
  switch (status) {
    case 1:
      return '售票中'
    case 2:
      return '已下架'
    case 3:
      return '已结束'
    default:
      return '待上架'
  }
}

export function canPurchaseEvent(detail?: ProgramEventDetail | null) {
  if (!detail || detail.status !== 1) {
    return false
  }

  const now = Date.now()
  const saleStart = parseDateTime(detail.saleStartTime)?.getTime() ?? 0
  const saleEnd = parseDateTime(detail.saleEndTime)?.getTime() ?? Number.MAX_SAFE_INTEGER
  return now >= saleStart && now <= saleEnd
}

export function deriveTicketTierState(
  tier: ProgramTicketTierInfo,
  detail?: ProgramEventDetail | null,
): TicketState {
  if (!detail || detail.status !== 1) {
    return {
      disabled: true,
      label: getEventStatusLabel(detail?.status),
      tagType: 'info',
    }
  }

  if (!canPurchaseEvent(detail)) {
    return {
      disabled: true,
      label: '未开售',
      tagType: 'warning',
    }
  }

  if (tier.status === 2 || tier.remainStock <= 0) {
    return {
      disabled: true,
      label: '已售罄',
      tagType: 'danger',
    }
  }

  if (tier.status !== 1) {
    return {
      disabled: true,
      label: '暂停售票',
      tagType: 'info',
    }
  }

  return {
    disabled: false,
    label: '可购买',
    tagType: 'success',
  }
}

export function buildNoticeList(detail: ProgramEventDetail) {
  return [
    detail.needRealName === 1
      ? '本场活动实行实名制入场，请携带下单时使用的有效证件。'
      : '本场活动默认非实名购票，具体以现场检票要求为准。',
    `每笔订单最多可购买 ${detail.purchaseLimit || 1} 张票。`,
    detail.ticketType === 2
      ? '当前活动支持纸质票配送，下单前请确认收货信息准确无误。'
      : '当前活动默认使用电子票，请在开演前于票夹中查看入场凭证。',
    `售票时间：${formatEventDetailTime(detail.saleStartTime, detail.saleEndTime)}。`,
  ]
}

export function buildEventCardDescription(event: ProgramEventBrief) {
  return `${event.city} | ${event.venueName || '场馆待定'}`
}

export function buildKeywordTitle(keyword: string, kind: ProgramPageKind) {
  const baseTitle = getPageTitle(kind)
  if (!keyword) {
    return baseTitle
  }
  return `“${keyword}”的搜索结果`
}
