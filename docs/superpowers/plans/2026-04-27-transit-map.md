# 跨運具整合地圖 (Transit Map) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在台北城市儀表板新增環狀線、機捷即時到站整合地圖與到站資訊板，以黑客松 PoC 品質可 Demo 為目標。

**Architecture:** Hybrid 方案 — 站點靜態資料由 Airflow DAG 一次性 seed 至 PostgreSQL，BE 讀取站點後對外暴露 `/api/v1/transit/stations`；即時到站時刻由 BE 直接 proxy TDX LiveBoard API 並回傳 `/api/v1/transit/arrivals`；FE 新增兩個 dashboard 卡片組件，開發期間使用 mock JSON，等 BE 就緒後換接真實 API。

**Tech Stack:**
- DE: Python, Apache Airflow, TDX API, pandas, SQLAlchemy, PostgreSQL
- BE: Go 1.21+, Gin, GORM, PostgreSQL (DBDashboard)
- FE: Vue 3, Pinia, Mapbox GL JS, Axios

---

> ⚠️ **三個分支可完全並行。** 各 Section 獨立執行，FE 用 mock 資料、BE 可用空 DB 先跑。

---

## 檔案結構

### feature/transit-de

```
Taipei-City-Dashboard-DE/dags/proj_city_dashboard/
├── circular_line_stations/
│   ├── __init__.py                         # 空檔
│   ├── job_config.json                     # DAG 元資料
│   └── circular_line_stations.py          # ETL 主程式
└── airport_mrt_stations/
    ├── __init__.py                         # 空檔
    ├── job_config.json                     # DAG 元資料
    └── airport_mrt_stations.py            # ETL 主程式
```

### feature/transit-be

```
Taipei-City-Dashboard-BE/app/
├── models/transit.go                       # TransitStation struct + GetStations()
├── controllers/transit.go                  # GetStations handler, GetArrivals handler
├── services/tdx.go                         # TDX OAuth token cache + ProxyLiveBoard()
└── routes/router.go                        # 修改：新增 configureTransitRoutes()
```

### feature/transit-fe

```
Taipei-City-Dashboard-FE/src/
├── assets/mock/
│   ├── transit-stations.json              # 站點 mock 資料
│   └── transit-arrivals.json             # 到站 mock 資料
├── store/transitStore.js                  # Pinia store：站點、到站、選取站點
└── dashboardComponent/components/
    ├── TransitMap.vue                     # 地圖主體（內嵌 Mapbox mini-map）
    └── ArrivalBoard.vue                   # 到站資訊列表
```

---

## Section A：feature/transit-de

### Task 1：建立 git worktree 與 DB table

**Files:**
- Create: `Taipei-City-Dashboard-DE/dags/proj_city_dashboard/circular_line_stations/__init__.py`
- Create: `Taipei-City-Dashboard-DE/dags/proj_city_dashboard/airport_mrt_stations/__init__.py`

- [ ] **Step 1：建立 worktree**

```bash
cd /home/nelson/Taipei-City-Dashboard
git worktree add ../transit-de feature/transit-de 2>/dev/null \
  || (git branch feature/transit-de && git worktree add ../transit-de feature/transit-de)
cd ../transit-de
```

- [ ] **Step 2：在 PostgreSQL (DBDashboard) 建立 transit_stations table**

連線至 Dashboard DB，執行：

```sql
CREATE TABLE IF NOT EXISTS transit_stations (
    id              SERIAL PRIMARY KEY,
    system          VARCHAR(20) NOT NULL,
    line_id         VARCHAR(20),
    station_id      VARCHAR(20) NOT NULL,
    station_name_zh VARCHAR(100),
    station_name_en VARCHAR(100),
    lat             DOUBLE PRECISION,
    lng             DOUBLE PRECISION,
    sequence        INT,
    data_time       TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_transit_stations_system_station
    ON transit_stations (system, station_id);
```

驗證：`SELECT COUNT(*) FROM transit_stations;` → 回傳 `0` 即成功。

- [ ] **Step 3：建立空的 `__init__.py`**

```bash
touch Taipei-City-Dashboard-DE/dags/proj_city_dashboard/circular_line_stations/__init__.py
touch Taipei-City-Dashboard-DE/dags/proj_city_dashboard/airport_mrt_stations/__init__.py
```

- [ ] **Step 4：設定 Airflow Variables（TDX 憑證）**

在 Airflow UI → Admin → Variables 新增：
- Key: `TDX_CLIENT_ID`，Value: `nelson40514-bc78d0c5-d6fb-4015`
- Key: `TDX_CLIENT_SECRET`，Value: `bef1185d-f06a-416e-aa8d-dc8cb90b2ab7`

或使用 CLI：
```bash
airflow variables set TDX_CLIENT_ID "nelson40514-bc78d0c5-d6fb-4015"
airflow variables set TDX_CLIENT_SECRET "bef1185d-f06a-416e-aa8d-dc8cb90b2ab7"
```

- [ ] **Step 5：Commit**

```bash
git add Taipei-City-Dashboard-DE/dags/proj_city_dashboard/circular_line_stations/__init__.py \
        Taipei-City-Dashboard-DE/dags/proj_city_dashboard/airport_mrt_stations/__init__.py
git commit -m "feat(de): scaffold transit_stations worktree and DB table"
```

---

### Task 2：環狀線站點 DAG（circular_line_stations）

**Files:**
- Create: `Taipei-City-Dashboard-DE/dags/proj_city_dashboard/circular_line_stations/job_config.json`
- Create: `Taipei-City-Dashboard-DE/dags/proj_city_dashboard/circular_line_stations/circular_line_stations.py`

- [ ] **Step 1：撰寫 job_config.json**

```json
{
    "dag_infos": {
        "dag_id": "circular_line_stations",
        "start_date": "2026-04-27",
        "schedule_interval": "@weekly",
        "catchup": false,
        "tags": ["transit_stations", "ntmetro", "環狀線", "TDX"],
        "description": "Seed New Taipei Metro (Circle Line) station locations from TDX.",
        "default_args": {
            "owner": "airflow",
            "email": ["DEFAULT_EMAIL_LIST"],
            "email_on_retry": false,
            "email_on_failure": true,
            "retries": 1,
            "retry_delay": 60
        },
        "ready_data_db": "postgres_default",
        "ready_data_default_table": "transit_stations",
        "ready_data_history_table": "",
        "raw_data_db": "postgres_default",
        "raw_data_table": "",
        "load_behavior": "replace"
    },
    "data_infos": {
        "name_cn": "新北捷運環狀線站點",
        "airflow_update_freq": "weekly",
        "source": "https://tdx.transportdata.tw/api/basic/v2/Rail/Metro/Station/NTMETRO",
        "source_type": "TDX API",
        "source_dept": "交通部 TDX",
        "gis_format": "Point",
        "output_coordinate": "EPSG:4326",
        "is_geometry": 0,
        "dataset_description": "新北捷運環狀線各站點位置與順序",
        "etl_description": "parse JSON, extract nested fields, standardize columns",
        "sensitivity": "public"
    }
}
```

- [ ] **Step 2：撰寫 ETL 主程式**

```python
# circular_line_stations.py
from airflow import DAG
from operators.common_pipeline import CommonDag


def _circular_line_stations(**kwargs):
    import pandas as pd
    from sqlalchemy import create_engine, text
    from utils.extract_stage import get_tdx_data
    from utils.load_stage import save_dataframe_to_postgresql
    from utils.transform_time import convert_str_to_time_format

    # Config
    ready_data_db_uri = kwargs.get("ready_data_db_uri")
    dag_infos = kwargs.get("dag_infos")
    load_behavior = dag_infos.get("load_behavior")
    default_table = dag_infos.get("ready_data_default_table")

    SYSTEM = "ntmetro"
    URL = "https://tdx.transportdata.tw/api/basic/v2/Rail/Metro/Station/NTMETRO?$format=JSON"

    # Extract
    raw_data = get_tdx_data(URL, output_format="dataframe")

    # Transform
    data = raw_data.copy()
    data["system"] = SYSTEM
    data["station_id"] = data["StationID"].astype(str)
    data["station_name_zh"] = data["StationName"].apply(
        lambda x: x.get("Zh_tw", "") if isinstance(x, dict) else ""
    )
    data["station_name_en"] = data["StationName"].apply(
        lambda x: x.get("En", "") if isinstance(x, dict) else ""
    )
    data["lat"] = data["StationPosition"].apply(
        lambda x: x.get("PositionLat", None) if isinstance(x, dict) else None
    )
    data["lng"] = data["StationPosition"].apply(
        lambda x: x.get("PositionLon", None) if isinstance(x, dict) else None
    )
    data["line_id"] = data.get("LineID", pd.Series(["YL"] * len(data)))
    data["sequence"] = data.get("StationSequence", pd.Series(range(len(data))))
    data["data_time"] = pd.Timestamp.now(tz="Asia/Taipei")

    ready_data = data[[
        "system", "line_id", "station_id",
        "station_name_zh", "station_name_en",
        "lat", "lng", "sequence", "data_time"
    ]]

    # Load — replace only ntmetro rows
    engine = create_engine(ready_data_db_uri)
    with engine.connect() as conn:
        conn.execute(
            text("DELETE FROM transit_stations WHERE system = :sys"),
            {"sys": SYSTEM}
        )
        conn.commit()
    ready_data.to_sql(
        default_table, engine, if_exists="append", index=False, schema="public"
    )


dag = CommonDag(
    proj_folder="proj_city_dashboard", dag_folder="circular_line_stations"
)
dag.create_dag(etl_func=_circular_line_stations)
```

- [ ] **Step 3：手動觸發 DAG 驗證**

在 Airflow UI 找到 `circular_line_stations`，點 Trigger DAG。

等待完成後，在 DB 執行：
```sql
SELECT system, station_id, station_name_zh, lat, lng
FROM transit_stations
WHERE system = 'ntmetro'
LIMIT 5;
```
預期：回傳 5 筆環狀線站點資料，lat/lng 不為 NULL。

- [ ] **Step 4：Commit**

```bash
git add Taipei-City-Dashboard-DE/dags/proj_city_dashboard/circular_line_stations/
git commit -m "feat(de): add circular_line_stations DAG for NTMETRO seed"
```

---

### Task 3：機捷站點 DAG（airport_mrt_stations）

**Files:**
- Create: `Taipei-City-Dashboard-DE/dags/proj_city_dashboard/airport_mrt_stations/job_config.json`
- Create: `Taipei-City-Dashboard-DE/dags/proj_city_dashboard/airport_mrt_stations/airport_mrt_stations.py`

- [ ] **Step 1：撰寫 job_config.json**

```json
{
    "dag_infos": {
        "dag_id": "airport_mrt_stations",
        "start_date": "2026-04-27",
        "schedule_interval": "@weekly",
        "catchup": false,
        "tags": ["transit_stations", "tymc", "機捷", "TDX"],
        "description": "Seed Taoyuan Airport MRT station locations from TDX.",
        "default_args": {
            "owner": "airflow",
            "email": ["DEFAULT_EMAIL_LIST"],
            "email_on_retry": false,
            "email_on_failure": true,
            "retries": 1,
            "retry_delay": 60
        },
        "ready_data_db": "postgres_default",
        "ready_data_default_table": "transit_stations",
        "ready_data_history_table": "",
        "raw_data_db": "postgres_default",
        "raw_data_table": "",
        "load_behavior": "replace"
    },
    "data_infos": {
        "name_cn": "桃園機場捷運站點",
        "airflow_update_freq": "weekly",
        "source": "https://tdx.transportdata.tw/api/basic/v2/Rail/Metro/Station/TYMC",
        "source_type": "TDX API",
        "source_dept": "交通部 TDX",
        "gis_format": "Point",
        "output_coordinate": "EPSG:4326",
        "is_geometry": 0,
        "dataset_description": "桃園機場捷運各站點位置與順序",
        "etl_description": "parse JSON, extract nested fields, standardize columns",
        "sensitivity": "public"
    }
}
```

- [ ] **Step 2：撰寫 ETL 主程式**

```python
# airport_mrt_stations.py
from airflow import DAG
from operators.common_pipeline import CommonDag


def _airport_mrt_stations(**kwargs):
    import pandas as pd
    from sqlalchemy import create_engine, text
    from utils.extract_stage import get_tdx_data

    # Config
    ready_data_db_uri = kwargs.get("ready_data_db_uri")
    dag_infos = kwargs.get("dag_infos")
    default_table = dag_infos.get("ready_data_default_table")

    SYSTEM = "tymc"
    URL = "https://tdx.transportdata.tw/api/basic/v2/Rail/Metro/Station/TYMC?$format=JSON"

    # Extract
    raw_data = get_tdx_data(URL, output_format="dataframe")

    # Transform
    data = raw_data.copy()
    data["system"] = SYSTEM
    data["station_id"] = data["StationID"].astype(str)
    data["station_name_zh"] = data["StationName"].apply(
        lambda x: x.get("Zh_tw", "") if isinstance(x, dict) else ""
    )
    data["station_name_en"] = data["StationName"].apply(
        lambda x: x.get("En", "") if isinstance(x, dict) else ""
    )
    data["lat"] = data["StationPosition"].apply(
        lambda x: x.get("PositionLat", None) if isinstance(x, dict) else None
    )
    data["lng"] = data["StationPosition"].apply(
        lambda x: x.get("PositionLon", None) if isinstance(x, dict) else None
    )
    data["line_id"] = data.get("LineID", pd.Series(["AL"] * len(data)))
    data["sequence"] = data.get("StationSequence", pd.Series(range(len(data))))
    data["data_time"] = pd.Timestamp.now(tz="Asia/Taipei")

    ready_data = data[[
        "system", "line_id", "station_id",
        "station_name_zh", "station_name_en",
        "lat", "lng", "sequence", "data_time"
    ]]

    # Load — replace only tymc rows
    engine = create_engine(ready_data_db_uri)
    with engine.connect() as conn:
        conn.execute(
            text("DELETE FROM transit_stations WHERE system = :sys"),
            {"sys": SYSTEM}
        )
        conn.commit()
    ready_data.to_sql(
        default_table, engine, if_exists="append", index=False, schema="public"
    )


dag = CommonDag(
    proj_folder="proj_city_dashboard", dag_folder="airport_mrt_stations"
)
dag.create_dag(etl_func=_airport_mrt_stations)
```

- [ ] **Step 3：手動觸發 DAG 驗證**

觸發 `airport_mrt_stations` DAG 後：
```sql
SELECT system, station_id, station_name_zh, lat, lng
FROM transit_stations
WHERE system = 'tymc'
LIMIT 5;
```
預期：回傳 5 筆機捷站點，lat/lng 不為 NULL。

總筆數驗證：
```sql
SELECT system, COUNT(*) FROM transit_stations GROUP BY system;
```
預期：`ntmetro` 和 `tymc` 各有合理筆數（ntmetro ≈ 51，tymc ≈ 21）。

- [ ] **Step 4：Commit**

```bash
git add Taipei-City-Dashboard-DE/dags/proj_city_dashboard/airport_mrt_stations/
git commit -m "feat(de): add airport_mrt_stations DAG for TYMC seed"
```

---

## Section B：feature/transit-be

### Task 4：建立 worktree 與 TransitStation model

**Files:**
- Create: `Taipei-City-Dashboard-BE/app/models/transit.go`

- [ ] **Step 1：建立 worktree**

```bash
cd /home/nelson/Taipei-City-Dashboard
git worktree add ../transit-be feature/transit-be 2>/dev/null \
  || (git branch feature/transit-be && git worktree add ../transit-be feature/transit-be)
cd ../transit-be
```

- [ ] **Step 2：撰寫 models/transit.go**

```go
// app/models/transit.go
package models

import "time"

// TransitStation maps to the transit_stations table in DBDashboard.
type TransitStation struct {
	ID            int       `gorm:"column:id;primaryKey" json:"id"`
	System        string    `gorm:"column:system" json:"system"`
	LineID        string    `gorm:"column:line_id" json:"line_id"`
	StationID     string    `gorm:"column:station_id" json:"station_id"`
	StationNameZh string    `gorm:"column:station_name_zh" json:"station_name_zh"`
	StationNameEn string    `gorm:"column:station_name_en" json:"station_name_en"`
	Lat           float64   `gorm:"column:lat" json:"lat"`
	Lng           float64   `gorm:"column:lng" json:"lng"`
	Sequence      int       `gorm:"column:sequence" json:"sequence"`
	DataTime      time.Time `gorm:"column:data_time" json:"data_time"`
}

func (TransitStation) TableName() string {
	return "transit_stations"
}

// GetTransitStations returns stations filtered by systems.
// systems is a slice of system codes, e.g. ["ntmetro", "tymc"].
// If systems is empty, all stations are returned.
func GetTransitStations(systems []string) ([]TransitStation, error) {
	var stations []TransitStation
	db := DBDashboard.Table("transit_stations")
	if len(systems) > 0 {
		db = db.Where("system IN ?", systems)
	}
	err := db.Order("system, sequence").Find(&stations).Error
	return stations, err
}
```

- [ ] **Step 3：驗證 model 可編譯**

```bash
cd Taipei-City-Dashboard-BE
go build ./app/models/...
```
預期：無錯誤輸出。

- [ ] **Step 4：Commit**

```bash
git add Taipei-City-Dashboard-BE/app/models/transit.go
git commit -m "feat(be): add TransitStation model and GetTransitStations query"
```

---

### Task 5：TDX Proxy service（services/tdx.go）

**Files:**
- Create: `Taipei-City-Dashboard-BE/app/services/tdx.go`

- [ ] **Step 1：撰寫 services/tdx.go**

```go
// app/services/tdx.go
package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

const tdxTokenURL = "https://tdx.transportdata.tw/auth/realms/TDXConnect/protocol/openid-connect/token"

type tdxTokenCache struct {
	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

var tdxCache = &tdxTokenCache{}

// getTDXToken returns a valid TDX bearer token, refreshing if expired.
func getTDXToken() (string, error) {
	tdxCache.mu.Lock()
	defer tdxCache.mu.Unlock()

	if time.Now().Before(tdxCache.expiresAt.Add(-60 * time.Second)) {
		return tdxCache.accessToken, nil
	}

	clientID := os.Getenv("TDX_CLIENT_ID")
	clientSecret := os.Getenv("TDX_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("TDX_CLIENT_ID or TDX_CLIENT_SECRET not set")
	}

	resp, err := http.PostForm(tdxTokenURL, map[string][]string{
		"grant_type":    {"client_credentials"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	})
	if err != nil {
		return "", fmt.Errorf("TDX token request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("TDX token decode failed: %w", err)
	}

	tdxCache.accessToken = result.AccessToken
	tdxCache.expiresAt = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	return tdxCache.accessToken, nil
}

// FetchTDXLiveBoard calls the TDX LiveBoard API for the given operator
// and returns the raw JSON bytes.
// operator is one of: "NTMETRO", "TYMC"
func FetchTDXLiveBoard(operator string) ([]byte, error) {
	token, err := getTDXToken()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"https://tdx.transportdata.tw/api/basic/v2/Rail/Metro/LiveBoard/%s?$format=JSON",
		operator,
	)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("TDX LiveBoard request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TDX LiveBoard returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
```

- [ ] **Step 2：驗證可編譯**

```bash
cd Taipei-City-Dashboard-BE
go build ./app/services/...
```
預期：無錯誤。

- [ ] **Step 3：Commit**

```bash
git add Taipei-City-Dashboard-BE/app/services/tdx.go
git commit -m "feat(be): add TDX OAuth token cache and LiveBoard proxy service"
```

---

### Task 6：Transit controllers（controllers/transit.go）

**Files:**
- Create: `Taipei-City-Dashboard-BE/app/controllers/transit.go`

- [ ] **Step 1：撰寫 controllers/transit.go**

```go
// app/controllers/transit.go
package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"TaipeiCityDashboardBE/app/models"
	"TaipeiCityDashboardBE/app/services"

	"github.com/gin-gonic/gin"
)

/*
GetTransitStations returns transit station list.
GET /api/v1/transit/stations?system=ntmetro,tymc
*/
func GetTransitStations(c *gin.Context) {
	systemParam := c.Query("system")
	var systems []string
	if systemParam != "" {
		for _, s := range strings.Split(systemParam, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				systems = append(systems, s)
			}
		}
	}

	stations, err := models.GetTransitStations(systems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stations})
}

/*
GetTransitArrivals proxies TDX LiveBoard API.
GET /api/v1/transit/arrivals?system=ntmetro&station_id=YL01
*/
func GetTransitArrivals(c *gin.Context) {
	system := c.Query("system")
	if system == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "system parameter is required (ntmetro or tymc)",
		})
		return
	}

	operatorMap := map[string]string{
		"ntmetro": "NTMETRO",
		"tymc":    "TYMC",
	}
	operator, ok := operatorMap[strings.ToLower(system)]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "system must be ntmetro or tymc",
		})
		return
	}

	rawBytes, err := services.FetchTDXLiveBoard(operator)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "transit data temporarily unavailable",
		})
		return
	}

	// Parse TDX response and normalize
	var tdxItems []map[string]interface{}
	if err := json.Unmarshal(rawBytes, &tdxItems); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "failed to parse TDX response",
		})
		return
	}

	stationFilter := strings.ToUpper(c.Query("station_id"))

	type Arrival struct {
		StationID     string `json:"station_id"`
		StationNameZh string `json:"station_name_zh"`
		Direction     int    `json:"direction"`
		DestinationZh string `json:"destination_zh"`
		EstimateTime  int    `json:"estimate_time"`
		UpdatedAt     string `json:"updated_at"`
	}

	var arrivals []Arrival
	for _, item := range tdxItems {
		sid, _ := item["StationID"].(string)
		if stationFilter != "" && strings.ToUpper(sid) != stationFilter {
			continue
		}
		nameZh := ""
		if n, ok := item["StationName"].(map[string]interface{}); ok {
			nameZh, _ = n["Zh_tw"].(string)
		}
		dir := 0
		if d, ok := item["TripHeadSign"].(map[string]interface{}); ok {
			_ = d
		}
		if d, ok := item["Direction"].(float64); ok {
			dir = int(d)
		}
		destZh := ""
		if d, ok := item["DestinationStationName"].(map[string]interface{}); ok {
			destZh, _ = d["Zh_tw"].(string)
		}
		est := 0
		if e, ok := item["EstimateTime"].(float64); ok {
			est = int(e)
		}
		updatedAt, _ := item["UpdateTime"].(string)

		arrivals = append(arrivals, Arrival{
			StationID:     sid,
			StationNameZh: nameZh,
			Direction:     dir,
			DestinationZh: destZh,
			EstimateTime:  est,
			UpdatedAt:     updatedAt,
		})
	}

	if arrivals == nil {
		arrivals = []Arrival{}
	}
	c.JSON(http.StatusOK, gin.H{"data": arrivals})
}
```

- [ ] **Step 2：驗證可編譯**

```bash
cd Taipei-City-Dashboard-BE
go build ./app/controllers/...
```
預期：無錯誤。

- [ ] **Step 3：Commit**

```bash
git add Taipei-City-Dashboard-BE/app/controllers/transit.go
git commit -m "feat(be): add GetTransitStations and GetTransitArrivals controllers"
```

---

### Task 7：注冊路由（routes/router.go）

**Files:**
- Modify: `Taipei-City-Dashboard-BE/app/routes/router.go`

- [ ] **Step 1：在 router.go 新增 transit 路由**

在 `ConfigureRoutes()` 函數的最後一行 `configureAIRoutes()` 之後加入：

```go
configureTransitRoutes()
```

在檔案最底部加入新函數：

```go
func configureTransitRoutes() {
	transitRoutes := RouterGroup.Group("/transit")
	transitRoutes.GET("/stations", controllers.GetTransitStations)
	transitRoutes.GET("/arrivals", controllers.GetTransitArrivals)
}
```

- [ ] **Step 2：確認 docker/.env 有 TDX 憑證**

在 `docker/.env` 新增（若不存在則建立）：
```
TDX_CLIENT_ID=nelson40514-bc78d0c5-d6fb-4015
TDX_CLIENT_SECRET=bef1185d-f06a-416e-aa8d-dc8cb90b2ab7
```

- [ ] **Step 3：Build 整個 BE 確認無錯**

```bash
cd Taipei-City-Dashboard-BE
go build ./...
```
預期：無錯誤。

- [ ] **Step 4：啟動 BE 並測試 stations endpoint**

```bash
# 啟動（視專案 docker 設定而定）
docker-compose up -d

# 測試（需先有 DB 資料，若 DE 尚未完成可跳過）
curl -s "http://localhost:8888/api/v1/transit/stations?system=ntmetro" | python3 -m json.tool | head -20
```
預期：回傳 `{"data": [...]}` 結構，或空陣列（DB 尚無資料時）。

- [ ] **Step 5：測試 arrivals endpoint**

```bash
curl -s "http://localhost:8888/api/v1/transit/arrivals?system=ntmetro" | python3 -m json.tool | head -20
```
預期：回傳 `{"data": [...]}` 結構（TDX 即時資料）。

- [ ] **Step 6：Commit**

```bash
git add Taipei-City-Dashboard-BE/app/routes/router.go
git commit -m "feat(be): register transit routes /stations and /arrivals"
```

---

## Section C：feature/transit-fe（供 Gemini 開發）

> **給 Gemini 的說明：** 本 Section 在 `feature/transit-fe` worktree 開發。BE 尚未就緒時使用 mock JSON，API 格式已在下方定義。完成後向 Nelson 回報，由 Nelson merge 至 develop。

### Task 8：建立 worktree 與 mock 資料

**Files:**
- Create: `Taipei-City-Dashboard-FE/src/assets/mock/transit-stations.json`
- Create: `Taipei-City-Dashboard-FE/src/assets/mock/transit-arrivals.json`

- [ ] **Step 1：建立 worktree**

```bash
cd /home/nelson/Taipei-City-Dashboard
git worktree add ../transit-fe feature/transit-fe 2>/dev/null \
  || (git branch feature/transit-fe && git worktree add ../transit-fe feature/transit-fe)
cd ../transit-fe
```

- [ ] **Step 2：建立站點 mock JSON**

```json
// Taipei-City-Dashboard-FE/src/assets/mock/transit-stations.json
{
  "data": [
    { "system": "ntmetro", "line_id": "YL", "station_id": "YL01", "station_name_zh": "新北產業園區", "station_name_en": "New Taipei Industrial Park", "lat": 25.0197, "lng": 121.4664, "sequence": 1 },
    { "system": "ntmetro", "line_id": "YL", "station_id": "YL02", "station_name_zh": "幸福", "station_name_en": "Xinfu", "lat": 25.0261, "lng": 121.4807, "sequence": 2 },
    { "system": "ntmetro", "line_id": "YL", "station_id": "YL03", "station_name_zh": "頭前庄", "station_name_en": "Touqianzhuang", "lat": 25.0319, "lng": 121.4938, "sequence": 3 },
    { "system": "ntmetro", "line_id": "YL", "station_id": "YL04", "station_name_zh": "新埔民生", "station_name_en": "Xinpu Minsheng", "lat": 25.0401, "lng": 121.4880, "sequence": 4 },
    { "system": "ntmetro", "line_id": "YL", "station_id": "YL05", "station_name_zh": "板橋", "station_name_en": "Banqiao", "lat": 25.0143, "lng": 121.4627, "sequence": 5 },
    { "system": "tymc", "line_id": "AL", "station_id": "A01", "station_name_zh": "台北車站", "station_name_en": "Taipei Main Station", "lat": 25.0479, "lng": 121.5165, "sequence": 1 },
    { "system": "tymc", "line_id": "AL", "station_id": "A02", "station_name_zh": "三重", "station_name_en": "Sanchong", "lat": 25.0632, "lng": 121.4851, "sequence": 2 },
    { "system": "tymc", "line_id": "AL", "station_id": "A03", "station_name_zh": "新北產業園區", "station_name_en": "New Taipei Industrial Park", "lat": 25.0198, "lng": 121.4666, "sequence": 3 },
    { "system": "tymc", "line_id": "AL", "station_id": "A12", "station_name_zh": "機場第一航廈", "station_name_en": "Airport Terminal 1", "lat": 25.0775, "lng": 121.2334, "sequence": 12 },
    { "system": "tymc", "line_id": "AL", "station_id": "A13", "station_name_zh": "機場第二航廈", "station_name_en": "Airport Terminal 2", "lat": 25.0818, "lng": 121.2302, "sequence": 13 }
  ]
}
```

- [ ] **Step 3：建立到站 mock JSON**

```json
// Taipei-City-Dashboard-FE/src/assets/mock/transit-arrivals.json
{
  "data": [
    { "station_id": "YL01", "station_name_zh": "新北產業園區", "direction": 0, "destination_zh": "迴龍", "estimate_time": 3, "updated_at": "2026-04-27T10:42:00Z" },
    { "station_id": "YL02", "station_name_zh": "幸福", "direction": 0, "destination_zh": "迴龍", "estimate_time": 7, "updated_at": "2026-04-27T10:42:00Z" },
    { "station_id": "YL01", "station_name_zh": "新北產業園區", "direction": 1, "destination_zh": "新埔民生", "estimate_time": 5, "updated_at": "2026-04-27T10:42:00Z" },
    { "station_id": "A01", "station_name_zh": "台北車站", "direction": 0, "destination_zh": "桃園機場", "estimate_time": 12, "updated_at": "2026-04-27T10:42:00Z" },
    { "station_id": "A01", "station_name_zh": "台北車站", "direction": 1, "destination_zh": "台北車站", "estimate_time": 8, "updated_at": "2026-04-27T10:42:00Z" }
  ]
}
```

- [ ] **Step 4：Commit**

```bash
git add Taipei-City-Dashboard-FE/src/assets/mock/
git commit -m "feat(fe): add transit mock JSON for stations and arrivals"
```

---

### Task 9：Pinia Store（transitStore.js）

**Files:**
- Create: `Taipei-City-Dashboard-FE/src/store/transitStore.js`

- [ ] **Step 1：撰寫 transitStore.js**

```js
// src/store/transitStore.js
import { defineStore } from "pinia";
import { ref } from "vue";
import http from "../router/axios.js";

// Toggle to true when BE is ready
const USE_MOCK = true;

import mockStations from "../assets/mock/transit-stations.json";
import mockArrivals from "../assets/mock/transit-arrivals.json";

export const useTransitStore = defineStore("transit", () => {
  const stations = ref([]);
  const arrivals = ref([]);
  const selectedStation = ref(null); // { station_id, system }
  const loading = ref(false);
  const error = ref(null);

  async function fetchStations(systems = []) {
    loading.value = true;
    error.value = null;
    try {
      if (USE_MOCK) {
        const all = mockStations.data;
        stations.value = systems.length
          ? all.filter((s) => systems.includes(s.system))
          : all;
      } else {
        const params = systems.length ? { system: systems.join(",") } : {};
        const res = await http.get("/transit/stations", { params });
        stations.value = res.data.data;
      }
    } catch (e) {
      error.value = e.message;
    } finally {
      loading.value = false;
    }
  }

  async function fetchArrivals(system, stationId = null) {
    loading.value = true;
    error.value = null;
    try {
      if (USE_MOCK) {
        const all = mockArrivals.data;
        arrivals.value = stationId
          ? all.filter((a) => a.station_id === stationId)
          : all.filter((a) => {
              const st = stations.value.find((s) => s.station_id === a.station_id);
              return st?.system === system;
            });
      } else {
        const params = { system };
        if (stationId) params.station_id = stationId;
        const res = await http.get("/transit/arrivals", { params });
        arrivals.value = res.data.data;
      }
    } catch (e) {
      error.value = e.message;
    } finally {
      loading.value = false;
    }
  }

  function selectStation(stationId, system) {
    selectedStation.value = stationId ? { station_id: stationId, system } : null;
    if (stationId) fetchArrivals(system, stationId);
  }

  return { stations, arrivals, selectedStation, loading, error, fetchStations, fetchArrivals, selectStation };
});
```

- [ ] **Step 2：確認無 lint 錯誤**

```bash
cd Taipei-City-Dashboard-FE
npx eslint src/store/transitStore.js --max-warnings=0
```
預期：無錯誤。

- [ ] **Step 3：Commit**

```bash
git add Taipei-City-Dashboard-FE/src/store/transitStore.js
git commit -m "feat(fe): add transitStore with mock/real toggle for stations and arrivals"
```

---

### Task 10：TransitMap.vue

**Files:**
- Create: `Taipei-City-Dashboard-FE/src/dashboardComponent/components/TransitMap.vue`

- [ ] **Step 1：撰寫 TransitMap.vue**

```vue
<!-- src/dashboardComponent/components/TransitMap.vue -->
<template>
  <div class="transit-map-card">
    <!-- System Toggle -->
    <div class="transit-map-card__toggles">
      <button
        v-for="sys in systemOptions"
        :key="sys.key"
        :class="['toggle-btn', { active: activeSystems.includes(sys.key) }]"
        :style="{ '--color': sys.color }"
        @click="toggleSystem(sys.key)"
      >
        {{ sys.label }}
      </button>
    </div>

    <!-- Map Container -->
    <div ref="mapContainer" class="transit-map-card__map" />

    <!-- Selected Station Popup (DOM overlay) -->
    <div v-if="selectedStationInfo" class="transit-map-card__popup">
      <div class="popup-line" :style="{ background: selectedStationInfo.color }">
        {{ selectedStationInfo.lineLabel }}
      </div>
      <div class="popup-name">{{ selectedStationInfo.station_name_zh }}</div>
      <div class="popup-hint">點擊查看到站資訊 ↓</div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch, computed } from "vue";
import mapboxGl from "mapbox-gl";
import "mapbox-gl/dist/mapbox-gl.css";
import { useTransitStore } from "../../store/transitStore.js";

const props = defineProps({
  systems: { type: Array, default: () => ["ntmetro", "tymc"] },
});
const emit = defineEmits(["station-click"]);

const transitStore = useTransitStore();
const mapContainer = ref(null);
let map = null;

const SYSTEM_CONFIG = {
  ntmetro: { color: "#eab308", label: "環狀線" },
  tymc:    { color: "#7c3aed", label: "機捷"   },
};
const systemOptions = Object.entries(SYSTEM_CONFIG).map(([key, v]) => ({ key, ...v }));
const activeSystems = ref([...props.systems]);

const selectedStationInfo = ref(null);

onMounted(async () => {
  await transitStore.fetchStations(activeSystems.value);

  mapboxGl.accessToken = import.meta.env.VITE_MAPBOXGL_ACCESS_TOKEN;
  map = new mapboxGl.Map({
    container: mapContainer.value,
    style: "mapbox://styles/mapbox/dark-v11",
    center: [121.46, 25.02],
    zoom: 10,
  });

  map.on("load", () => {
    renderStations();
  });

  map.on("click", "transit-stations-circle", (e) => {
    const feature = e.features[0];
    const { station_id, system, station_name_zh } = feature.properties;
    selectedStationInfo.value = {
      station_id,
      system,
      station_name_zh,
      color: SYSTEM_CONFIG[system]?.color ?? "#fff",
      lineLabel: SYSTEM_CONFIG[system]?.label ?? system,
    };
    emit("station-click", station_id, system);
  });

  map.on("mouseenter", "transit-stations-circle", () => {
    map.getCanvas().style.cursor = "pointer";
  });
  map.on("mouseleave", "transit-stations-circle", () => {
    map.getCanvas().style.cursor = "";
  });
});

onUnmounted(() => {
  map?.remove();
});

function renderStations() {
  if (!map) return;
  const filtered = transitStore.stations.filter((s) =>
    activeSystems.value.includes(s.system)
  );

  const geojson = {
    type: "FeatureCollection",
    features: filtered.map((s) => ({
      type: "Feature",
      geometry: { type: "Point", coordinates: [s.lng, s.lat] },
      properties: {
        station_id: s.station_id,
        system: s.system,
        station_name_zh: s.station_name_zh,
        color: SYSTEM_CONFIG[s.system]?.color ?? "#ffffff",
      },
    })),
  };

  if (map.getSource("transit-stations")) {
    map.getSource("transit-stations").setData(geojson);
  } else {
    map.addSource("transit-stations", { type: "geojson", data: geojson });
    map.addLayer({
      id: "transit-stations-circle",
      type: "circle",
      source: "transit-stations",
      paint: {
        "circle-radius": 7,
        "circle-color": ["get", "color"],
        "circle-stroke-color": "#ffffff",
        "circle-stroke-width": 2,
      },
    });
  }
}

function toggleSystem(sys) {
  const idx = activeSystems.value.indexOf(sys);
  if (idx >= 0) {
    activeSystems.value.splice(idx, 1);
  } else {
    activeSystems.value.push(sys);
  }
  transitStore.fetchStations(activeSystems.value).then(renderStations);
}

watch(() => transitStore.stations, renderStations);
</script>

<style scoped>
.transit-map-card {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 320px;
}
.transit-map-card__toggles {
  display: flex;
  gap: 8px;
  padding: 8px 12px;
  background: #1e293b;
}
.toggle-btn {
  padding: 4px 12px;
  border-radius: 20px;
  border: 2px solid var(--color);
  background: transparent;
  color: var(--color);
  cursor: pointer;
  font-size: 12px;
  transition: background 0.2s;
}
.toggle-btn.active {
  background: var(--color);
  color: #0f172a;
  font-weight: bold;
}
.transit-map-card__map {
  flex: 1;
  min-height: 260px;
}
.transit-map-card__popup {
  position: absolute;
  bottom: 12px;
  left: 12px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 10px 14px;
  pointer-events: none;
}
.popup-line {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  color: #0f172a;
  font-weight: bold;
  margin-bottom: 4px;
}
.popup-name {
  color: #f1f5f9;
  font-size: 14px;
}
.popup-hint {
  color: #64748b;
  font-size: 11px;
  margin-top: 2px;
}
</style>
```

- [ ] **Step 2：確認無 lint 錯誤**

```bash
cd Taipei-City-Dashboard-FE
npx eslint src/dashboardComponent/components/TransitMap.vue --max-warnings=0
```

- [ ] **Step 3：Commit**

```bash
git add Taipei-City-Dashboard-FE/src/dashboardComponent/components/TransitMap.vue
git commit -m "feat(fe): add TransitMap.vue with Mapbox mini-map and system toggles"
```

---

### Task 11：ArrivalBoard.vue

**Files:**
- Create: `Taipei-City-Dashboard-FE/src/dashboardComponent/components/ArrivalBoard.vue`

- [ ] **Step 1：撰寫 ArrivalBoard.vue**

```vue
<!-- src/dashboardComponent/components/ArrivalBoard.vue -->
<template>
  <div class="arrival-board">
    <!-- System Selector -->
    <div class="arrival-board__header">
      <select v-model="selectedSystem" class="system-select">
        <option value="ntmetro">環狀線</option>
        <option value="tymc">機捷</option>
      </select>
      <span class="update-time">{{ lastUpdated }}</span>
    </div>

    <!-- Arrival List -->
    <div v-if="filteredArrivals.length > 0" class="arrival-board__list">
      <div
        v-for="(item, idx) in filteredArrivals"
        :key="idx"
        class="arrival-item"
        :style="{ '--line-color': systemColor }"
      >
        <div class="arrival-item__dot" />
        <div class="arrival-item__info">
          <div class="arrival-item__station">{{ item.station_name_zh }}</div>
          <div class="arrival-item__dir">往 {{ item.destination_zh }}</div>
        </div>
        <div class="arrival-item__time">
          {{ item.estimate_time }}<span class="unit">分</span>
        </div>
        <div class="arrival-item__id">{{ item.station_id }}</div>
      </div>
    </div>

    <div v-else class="arrival-board__empty">
      {{ transitStore.loading ? "載入中..." : "目前無到站資訊" }}
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { useTransitStore } from "../../store/transitStore.js";

const props = defineProps({
  system: { type: String, default: "ntmetro" },
  stationId: { type: String, default: null },
});

const transitStore = useTransitStore();
const selectedSystem = ref(props.system);
const lastUpdated = ref("");
let timer = null;

const SYSTEM_COLORS = { ntmetro: "#eab308", tymc: "#7c3aed" };
const systemColor = computed(() => SYSTEM_COLORS[selectedSystem.value] ?? "#ffffff");

const filteredArrivals = computed(() => {
  if (!transitStore.arrivals) return [];
  return transitStore.arrivals.slice(0, 8);
});

async function refresh() {
  await transitStore.fetchArrivals(selectedSystem.value, props.stationId);
  lastUpdated.value = `更新於 ${new Date().toLocaleTimeString("zh-TW", { hour: "2-digit", minute: "2-digit", second: "2-digit" })}`;
}

onMounted(() => {
  refresh();
  timer = setInterval(refresh, 30000);
});
onUnmounted(() => clearInterval(timer));

watch([selectedSystem, () => props.stationId], refresh);
watch(() => props.system, (v) => { selectedSystem.value = v; });
</script>

<style scoped>
.arrival-board {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #0f172a;
}
.arrival-board__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background: #1e293b;
}
.system-select {
  background: #0f172a;
  border: 1px solid #334155;
  color: #94a3b8;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
}
.update-time { font-size: 10px; color: #475569; }
.arrival-board__list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.arrival-item {
  display: grid;
  grid-template-columns: 12px 1fr auto auto;
  gap: 10px;
  align-items: center;
  background: #1e293b;
  padding: 10px 14px;
  border-radius: 6px;
  border-left: 3px solid var(--line-color);
}
.arrival-item__dot {
  width: 8px; height: 8px;
  background: var(--line-color);
  border-radius: 50%;
}
.arrival-item__station { color: #f1f5f9; font-size: 13px; }
.arrival-item__dir { color: #64748b; font-size: 11px; }
.arrival-item__time {
  color: #22d3ee;
  font-size: 20px;
  font-weight: bold;
  font-family: monospace;
}
.unit { font-size: 11px; color: #64748b; }
.arrival-item__id { color: #64748b; font-size: 10px; }
.arrival-board__empty {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #475569;
  font-size: 13px;
}
</style>
```

- [ ] **Step 2：確認無 lint 錯誤**

```bash
npx eslint src/dashboardComponent/components/ArrivalBoard.vue --max-warnings=0
```

- [ ] **Step 3：Commit**

```bash
git add Taipei-City-Dashboard-FE/src/dashboardComponent/components/ArrivalBoard.vue
git commit -m "feat(fe): add ArrivalBoard.vue with 30s auto-refresh"
```

---

### Task 12：串接 TransitMap → ArrivalBoard 並加入 DashboardView

**Files:**
- Modify: `Taipei-City-Dashboard-FE/src/views/DashboardView.vue`

> **架構說明：** DashboardView 是資料驅動架構（DB config → DashboardComponent）。PoC 採用硬編碼方式在頁面底部插入 Transit 區塊，最快可跑。

- [ ] **Step 1：在 DashboardView.vue 加入 imports**

在 `src/views/DashboardView.vue` 的 `<script setup>` 區塊中，找到現有的最後一個 import 後加入：

```js
import { ref } from "vue";
import TransitMap from "../dashboardComponent/components/TransitMap.vue";
import ArrivalBoard from "../dashboardComponent/components/ArrivalBoard.vue";

const arrivalSystem = ref("ntmetro");
const arrivalStationId = ref(null);

function onTransitStationClick(stationId, system) {
  arrivalSystem.value = system;
  arrivalStationId.value = stationId;
}
```

- [ ] **Step 2：在 template 加入 Transit 區塊**

在 `</template>` 結束標籤的**前一行**（現有 `<MoreInfo />` 和 `<ReportIssue />` 之後）加入：

```vue
<!-- Transit Map PoC Section -->
<div class="transit-poc-section">
  <h3 class="transit-poc-title">跨運具整合地圖</h3>
  <div class="transit-poc-grid">
    <div class="transit-poc-map-card">
      <TransitMap
        :systems="['ntmetro', 'tymc']"
        @station-click="onTransitStationClick"
      />
    </div>
    <div class="transit-poc-arrival-card">
      <ArrivalBoard
        :system="arrivalSystem"
        :station-id="arrivalStationId"
      />
    </div>
  </div>
</div>
```

- [ ] **Step 3：在 DashboardView.vue 加入 scoped 樣式**

在 `</template>` 後加入（若檔案已有 `<style>`，在其中加入以下 class）：

```vue
<style scoped>
.transit-poc-section {
  padding: 24px 16px;
  background: #0a0e1a;
}
.transit-poc-title {
  color: #94a3b8;
  font-size: 14px;
  text-transform: uppercase;
  letter-spacing: 1px;
  margin-bottom: 16px;
}
.transit-poc-grid {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 16px;
  min-height: 360px;
}
.transit-poc-map-card,
.transit-poc-arrival-card {
  background: #0f172a;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid #1e293b;
}
</style>
```

- [ ] **Step 4：啟動 dev server 目視確認**

```bash
npm run dev
```

打開 http://localhost:5173，向下滾動到底部的「跨運具整合地圖」區塊，確認：
- 地圖出現環狀線（黃色點）和機捷（紫色點）站點（共 10 個 mock 站點）
- 點擊站點後右側 ArrivalBoard 更新顯示該站到站資訊
- Toggle 按鈕可開關各系統顯示
- ArrivalBoard 右上角顯示最後更新時間

- [ ] **Step 5：Commit**

```bash
git add src/views/DashboardView.vue
git commit -m "feat(fe): embed TransitMap + ArrivalBoard in DashboardView as hardcoded PoC section"
```

---

### Task 13：切換至真實 API（BE ready 後執行）

**Files:**
- Modify: `Taipei-City-Dashboard-FE/src/store/transitStore.js`

- [ ] **Step 1：確認 BE endpoints 可用**

```bash
curl -s "http://<BE_HOST>/api/v1/transit/stations?system=ntmetro" | python3 -m json.tool | head -5
curl -s "http://<BE_HOST>/api/v1/transit/arrivals?system=ntmetro" | python3 -m json.tool | head -5
```
預期：兩個 endpoint 均回傳 `{"data": [...]}`。

- [ ] **Step 2：關閉 mock 模式**

在 `src/store/transitStore.js` 第 7 行：
```js
// 改為：
const USE_MOCK = false;
```

- [ ] **Step 3：啟動 dev server 確認真實資料**

```bash
npm run dev
```

確認 Network tab 中有呼叫 `/api/v1/transit/stations` 和 `/api/v1/transit/arrivals`，且地圖上的站點數量與 DB seed 資料一致。

- [ ] **Step 4：Commit**

```bash
git add Taipei-City-Dashboard-FE/src/store/transitStore.js
git commit -m "feat(fe): switch transitStore from mock to real BE API"
```

---

## 遷移備案記錄（Plan A）

下列步驟**不在本次實作範圍**，記錄供未來執行：

1. 建立 `transit_arrivals_history` table（含 `fetched_at TIMESTAMPTZ`）
2. 新增 DAG：`circular_line_arrivals_etl.py`（schedule `* * * * *`）
3. 新增 DAG：`airport_mrt_arrivals_etl.py`（schedule `* * * * *`）
4. `services/tdx.go` 的 `FetchTDXLiveBoard()` 改為從 DB 讀取
5. 評估 Airflow worker 資源（每分鐘 DAG 需 Celery beat 或獨立排程）

