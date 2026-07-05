/**
 * @bit-labs.cn/metrics-js - 轻量 Web 指标采集 SDK
 */
const STORAGE_ANON = 'metrics_anon_id'
const STORAGE_SESSION = 'metrics_session_id'

let config = {
  appKey: '',
  host: '',
  platform: 'web',
  autoPageView: true,
  flushIntervalMs: 3000,
  maxBatch: 20,
}

let queue = []
let flushTimer = null
let initialized = false

function normalizeHost(host) {
  return String(host || '').replace(/\/+$/, '')
}

function apiUrl() {
  return `${normalizeHost(config.host)}/api/v1/metrics/collect`
}

function uuid() {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) return crypto.randomUUID()
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, c => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

function getAnonId() {
  try {
    let id = localStorage.getItem(STORAGE_ANON)
    if (!id) {
      id = uuid()
      localStorage.setItem(STORAGE_ANON, id)
    }
    return id
  } catch {
    return uuid()
  }
}

function getSessionId() {
  try {
    let id = sessionStorage.getItem(STORAGE_SESSION)
    if (!id) {
      id = uuid()
      sessionStorage.setItem(STORAGE_SESSION, id)
    }
    return id
  } catch {
    return uuid()
  }
}

function baseEvent(name, props = {}) {
  return {
    appKey: config.appKey,
    name,
    anonId: getAnonId(),
    sessionId: getSessionId(),
    platform: config.platform,
    url: typeof location !== 'undefined' ? location.href : '',
    referrer: typeof document !== 'undefined' ? document.referrer : '',
    ua: typeof navigator !== 'undefined' ? navigator.userAgent : '',
    clientTs: Date.now(),
    props: props && typeof props === 'object' ? props : {},
  }
}

function scheduleFlush() {
  if (flushTimer) return
  flushTimer = setTimeout(() => {
    flushTimer = null
    flush()
  }, config.flushIntervalMs)
}

function sendPayload(body) {
  const url = apiUrl()
  const json = JSON.stringify(body)
  if (typeof navigator !== 'undefined' && navigator.sendBeacon) {
    const blob = new Blob([json], { type: 'application/json' })
    if (navigator.sendBeacon(url, blob)) return Promise.resolve()
  }
  return fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: json,
    keepalive: true,
  }).catch(() => {})
}

export function init(opts = {}) {
  config = {
    ...config,
    ...opts,
    host: normalizeHost(opts.host || config.host || (typeof location !== 'undefined' ? location.origin : '')),
  }
  if (!config.appKey) {
    console.warn('[Metrics] appKey is required')
    return
  }
  if (initialized) return
  initialized = true

  bindDeclarativeClicks()
  if (config.autoPageView) {
    track('page_view', { path: typeof location !== 'undefined' ? location.pathname : '' })
  }
  if (typeof window !== 'undefined') {
    window.addEventListener('beforeunload', () => flush(true))
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'hidden') flush(true)
    })
  }
}

export function track(name, props) {
  if (!config.appKey) return
  queue.push(baseEvent(name, props))
  if (queue.length >= config.maxBatch) {
    flush()
  } else {
    scheduleFlush()
  }
}

export function flush(sync = false) {
  if (!queue.length) return Promise.resolve()
  const batch = queue.splice(0, config.maxBatch)
  const body = batch.length === 1 ? batch[0] : { events: batch }
  if (sync) {
    sendPayload(body)
    return Promise.resolve()
  }
  return sendPayload(body)
}

function bindDeclarativeClicks() {
  if (typeof document === 'undefined') return
  document.addEventListener('click', e => {
    const el = e.target?.closest?.('[data-metrics-event]')
    if (!el) return
    const name = el.getAttribute('data-metrics-event') || 'click'
    const props = {}
    for (const attr of el.attributes) {
      if (attr.name.startsWith('data-metrics-') && attr.name !== 'data-metrics-event') {
        const key = attr.name.slice('data-metrics-'.length).replace(/-([a-z])/g, (_, c) => c.toUpperCase())
        props[key] = attr.value
      }
    }
    track(name, props)
  }, true)
}

export const Metrics = { init, track, flush }

if (typeof window !== 'undefined') {
  window.Metrics = Metrics
}
