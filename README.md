# owl-metrics（通用指标采集子系统）

独立 owl SubApp，提供自建埋点/分析能力：访问量、下载量、点击、菜单/功能使用等任意事件。

## 目录

```text
owl-metrics/
  app/                 # Go SubApp（model / repository / service / handle / route）
  sdk/js/              # Web JS SDK（也可复制到静态站点 js/metrics.js）
  sdk/go/              # Go SDK（桌面端 / 后端埋点）
  frontend/admin/      # @bit-labs.cn/owl-metrics-ui 管理看板
```

## 启动

在 [`bit-labs`](../bit-labs) 聚合启动（已注册 `SubAppMetrics`）：

```bash
cd ../bit-labs
go mod tidy
go run .
```

管理后台菜单：**指标分析** → 数据看板 / 应用管理。

## 接入流程

1. 后台 **指标分析 → 应用管理** 创建应用，获得 `AppKey` / `AppSecret`。
2. **Web**：`Metrics.init({ appKey, host })`，`Metrics.track('event', props)`。
3. **Go 桌面/后端**：使用 `bit-labs.cn/owl-metrics/sdk/go`，配置 `Endpoint`、`AppKey`、`AppSecret`（HMAC 签名）。

### Web 示例

```html
<script type="module">
  import { init, track } from './js/metrics.js'
  init({ appKey: 'your-app-key', host: location.origin })
  track('menu_click', { menu: 'ssh' })
</script>
```

声明式点击：`data-metrics-event="download" data-metrics-os="windows"`

### Go 示例

```go
c := metrics.New(metrics.Config{
  Endpoint:  "http://127.0.0.1:8080",
  AppKey:    "...",
  AppSecret: "...",
  Platform:  "desktop",
})
c.Track(ctx, metrics.Event{Name: "feature_use", Props: map[string]any{"feature": "ssh"}})
```

### Firefly 官网

在 `cloud/frontend/website/js/config.js` 设置 `METRICS_APP_KEY`（或 `window.__FIREFLY_METRICS_APP_KEY__`）。

### Firefly 桌面端

环境变量：

- `FIREFLY_METRICS_APP_KEY`
- `FIREFLY_METRICS_APP_SECRET`
- `FIREFLY_METRICS_ENDPOINT`（可选，默认使用 auth API 地址）
- `FIREFLY_METRICS_ENABLED=false` 可关闭

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/metrics/collect` | 公开采集（Web sendBeacon / Go HMAC） |
| GET | `/api/v1/metrics/pixel.gif` | 像素兜底 |
| CRUD | `/api/v1/metrics/admin/apps` | 应用管理 |
| GET | `/api/v1/metrics/admin/overview` | KPI 概览 |
| GET | `/api/v1/metrics/admin/timeseries` | 按天趋势 |
| GET | `/api/v1/metrics/admin/topn` | 维度 TopN |
| GET | `/api/v1/metrics/admin/events` | 事件浏览器 |

## Rollup

每日将明细聚合到 `metrics_rollup_daily`：

```bash
go run . metrics-rollup
```

启动时也会后台尝试聚合昨日数据。

## 存储扩展

`service.MetricsStore` 接口预留 ClickHouse 等实现；当前默认 `SQLMetricsStore`（PostgreSQL + GORM）。
