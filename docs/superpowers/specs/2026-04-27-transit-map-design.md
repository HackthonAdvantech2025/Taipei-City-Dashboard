# 跨運具整合地圖｜設計規格

**日期**：2026-04-27  
**範疇**：黑客松 PoC — Feature A（跨運具整合地圖）  
**狀態**：已核准，待實作

---

## 1. 目標

在台北城市儀表板現有 Dashboard 新增兩個組件，將環狀線（新北捷運）、桃園機捷、現有台北捷運、YouBike 整合至同一即時地圖，並附帶到站時刻板，提升多運具通勤情境的資訊密度。

**Demo 核心亮點**：四套運輸系統在同一地圖上，點擊站點即顯示即時到站倒數。

---

## 2. 範疇

### 本次實作（PoC）
- 環狀線站點資料 + 即時到站時刻
- 桃園機捷站點資料 + 即時到站時刻
- 跨運具整合地圖組件（`TransitMap.vue`）
- 即時到站資訊板組件（`ArrivalBoard.vue`）

### 明確排除（已記錄為未來 Roadmap）
- Feature B：氣候感應通勤優化器
- Feature C：YouBike 最後一哩路個人化

---

## 3. 架構（Hybrid 方案 B）

```
TDX API (Stations)
    │
    ▼ 一次性 seed
Airflow DAG ──────→ PostgreSQL (transit_stations)
                          │
                          ▼
TDX API (LiveBoard) ── BE Proxy ──→ /api/v1/transit/arrivals
                          │
PostgreSQL ───────────────┤
                          ▼
                       Vue FE
                  TransitMap.vue
                  ArrivalBoard.vue
```

**選擇理由**：即時到站資料每分鐘變動，Airflow 排程頻率不足；站點靜態資料透過一次性 DAG seed 持久化，兼顧資料管控與開發速度。

---

## 4. Git Worktree 分支規劃

| 分支 | 負責範圍 | 開發者 | 相依性 |
|------|---------|--------|--------|
| `feature/transit-de` | Airflow DAG seed | — | 無，可立即開始 |
| `feature/transit-be` | Go API + TDX Proxy | — | 可用 mock DB 先開發 |
| `feature/transit-fe` | Vue 組件 | Gemini | 可用 mock JSON 先開發 |

三個分支可**完全並行**開發。

---

## 5. DE 層（`feature/transit-de`）

### 新增 DAG

```
dags/proj_city_dashboard/
├── circular_line_stations/
│   ├── __init__.py
│   ├── job_config.json
│   └── circular_line_stations.py
└── airport_mrt_stations/
    ├── __init__.py
    ├── job_config.json
    └── airport_mrt_stations.py
```

**Schedule**：`@weekly`（站點資料不常變動）  
**TDX Operator codes**：`NTMETRO`（環狀線）、`TYMC`（機捷）  
**TDX Station API**：`GET /v2/Rail/Metro/Station/{operator}`

### DB Schema

```sql
CREATE TABLE transit_stations (
    id              SERIAL PRIMARY KEY,
    system          VARCHAR(20) NOT NULL,  -- 'ntmetro' | 'tymc'
    line_id         VARCHAR(20),
    station_id      VARCHAR(20) NOT NULL,
    station_name_zh VARCHAR(100),
    station_name_en VARCHAR(100),
    lat             DOUBLE PRECISION,
    lng             DOUBLE PRECISION,
    sequence        INT,
    data_time       TIMESTAMPTZ DEFAULT NOW()
);
```

**Load behavior**：`replace`（每次全量替換）

---

## 6. BE 層（`feature/transit-be`）

### 新增檔案

```
app/
├── models/transit.go       # TransitStation struct + DB query
├── controllers/transit.go  # HTTP handlers
├── services/tdx.go         # TDX OAuth token cache + proxy
└── routes/router.go        # 新增 configureTransitRoutes()
```

### API Endpoints

#### `GET /api/v1/transit/stations`

| 參數 | 型別 | 說明 |
|------|------|------|
| `system` | `string` | 逗號分隔，空白=全部。例：`ntmetro,tymc` |

**Response**：
```json
{
  "data": [
    {
      "system": "ntmetro",
      "line_id": "YL",
      "station_id": "YL01",
      "station_name_zh": "新北產業園區",
      "station_name_en": "New Taipei Industrial Park",
      "lat": 25.0123,
      "lng": 121.4567,
      "sequence": 1
    }
  ]
}
```

#### `GET /api/v1/transit/arrivals`

| 參數 | 型別 | 說明 |
|------|------|------|
| `system` | `string` | 必填。`ntmetro` 或 `tymc` |
| `station_id` | `string` | 選填，空白=全線 |

**Response**：
```json
{
  "data": [
    {
      "station_id": "YL01",
      "station_name_zh": "新北產業園區",
      "direction": 0,
      "destination_zh": "迴龍",
      "estimate_time": 3,
      "updated_at": "2026-04-27T10:30:00Z"
    }
  ]
}
```

### TDX Token 管理（`services/tdx.go`）

- 環境變數：`TDX_CLIENT_ID`、`TDX_CLIENT_SECRET`
- In-memory token cache，有效期 86400 秒，快到期自動 refresh
- Proxy 失敗：回 `503`，body `{"error": "transit data temporarily unavailable"}`
- TDX LiveBoard API：`GET /v2/Rail/Metro/LiveBoard/{operator}`

---

## 7. FE 層（`feature/transit-fe`）

### 新增組件

```
src/
├── dashboardComponent/
│   ├── TransitMap.vue        # 跨運具整合地圖
│   └── ArrivalBoard.vue      # 即時到站資訊板
└── store/
    └── transit.js            # Pinia store，管理站點與到站資料
```

### 組件介面

#### TransitMap.vue

```js
// Props
systems: String[]  // ['ntmetro', 'tymc', 'tmrt', 'youbike']，控制顯示哪些系統

// Emits
station-click(stationId: string, system: string)
```

**功能**：
- 地圖顯示各系統路線（顏色：黃=環狀線、紫=機捷、紅=台北捷運、綠=YouBike）
- 右上角系統 toggle，可開關各系統的顯示
- 點擊站點 → popup 顯示站名 + 最近到站時刻，同時 emit `station-click`

#### ArrivalBoard.vue

```js
// Props
system: String      // 'ntmetro' | 'tymc'
stationId: String | null  // null = 全線列表
```

**功能**：
- 顯示指定路線/站點的即時到站倒數（分鐘）
- 每 30 秒自動呼叫 `/api/v1/transit/arrivals` 更新
- 顯示最後更新時間

### Dashboard 配置

- `TransitMap` 佔 **2 欄寬**（現有 dashboard 3 欄格線）
- `ArrivalBoard` 佔 **1 欄寬**
- 串接方式：TransitMap `station-click` → 父層更新 ArrivalBoard 的 `stationId` prop

### 各系統站點資料來源（TransitMap 用）

| System | 資料來源 | 說明 |
|--------|---------|------|
| `ntmetro` | `GET /api/v1/transit/stations?system=ntmetro` | 本次新建 |
| `tymc` | `GET /api/v1/transit/stations?system=tymc` | 本次新建 |
| `youbike` | 現有 YouBike 站點 endpoint（`tran_ubike_station`） | 已存在，FE 直接串 |
| `tmrt` | 本次 PoC **不顯示台北捷運站點**，地圖上以路線裝飾線呈現即可 | 避免額外 DAG |

### Mock 資料（開發用）

FE 開發期間使用靜態 JSON（`src/assets/mock/transit-stations.json`、`transit-arrivals.json`），結構與 BE response 相同，等 BE 完成後替換 API endpoint。

---

## 8. 遷移備案（Plan A — 全套 Pipeline）

> 本節記錄 PoC 完成後遷移至完整 DE 架構的路徑，不在黑客松範疇內執行。

**目標**：將 BE proxy 改為 Airflow DAG 每分鐘抓取，寫入 DB，BE 改讀 DB。

**步驟**：
1. 新增 `transit_arrivals_history` table（含 `fetched_at TIMESTAMPTZ`）
2. 新增 DAG：`circular_line_arrivals_etl.py`、`airport_mrt_arrivals_etl.py`，schedule `* * * * *`（每分鐘）
3. 新增 DAG：`tdx_token_refresh.py`，schedule `@daily`，管理 TDX token 持久化
4. `services/tdx.go` 從 HTTP proxy 重構為 DB reader
5. `controllers/transit.go` 的 `GetArrivals` 切換資料源至 DB

**注意**：每分鐘執行的 DAG 在 Airflow 上需評估 worker 資源；可考慮改用 Celery beat 或獨立排程服務。

---

## 9. Future Roadmap

### Feature B：氣候感應通勤優化器

**觸發條件**：Feature A 上線後  
**資料基礎**：`cwb_city_weather_forecast`（D050102_1）+ 捷運歷史運量  
**實作路徑**：
1. 新 DAG `commute_weather_index_etl.py`（每 6 小時），產出雨天通勤指數
2. BE：`GET /api/v1/predict/commuting`
3. FE：`WeatherCommute.vue` 卡片

### Feature C：YouBike 最後一哩路個人化

**觸發條件**：Feature A 上線後  
**資料基礎**：`tran_ubike_realtime` + `traffic_youbike_two_realtime` 歷史累積  
**實作路徑**：
1. 離線分析：各站尖峰時段缺車/缺位概率
2. 新 table：`youbike_station_demand_profile`
3. BE：`GET /api/v1/youbike/recommendation?lat=&lng=&time=`
4. FE：`YouBikeLastMile.vue`

---

## 10. 技術決策記錄

| 決策 | 選擇 | 理由 |
|------|------|------|
| 即時資料來源 | BE Proxy TDX | Airflow 排程頻率不足應付每分鐘更新 |
| 站點資料持久化 | DAG seed → PostgreSQL | 保持資料在 DB 掌控，FE 不直接打外部 API |
| FE 開發分工 | Gemini 負責 | 三分支並行，FE 先用 mock JSON |
| 組件位置 | 現有 Dashboard 新增 card | 成本低，維持系統一致性 |
| 遷移策略 | 記錄但不執行 | 黑客松優先 demo，架構完整性非核心目標 |
