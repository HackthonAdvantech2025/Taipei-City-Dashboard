# 食安組件開發文件 (Food Safety Components Documentation)

## 0. 提案方向與內容 (Proposal Direction and Content)

### 0.1 提案啟發：打破資訊不對稱 (Inspiration)
食品安全問題一旦發生往往會引發大規模社會恐慌。然而，食安風險通常並非突發，而是長期衛生管理不善的累積。目前的困境在於**市民缺乏足夠透明、直觀的資訊**來了解周遭商家的食安狀況。
本提案希望透過**「公開透明、視覺化」**的儀表板，將繁瑣的政府稽查數據轉化為易讀的地圖與統計圖表，讓食安資訊不再是冷冰冰的公文，而是市民日常生活中隨手可得的參考指標。

### 0.2 對市民的影響：從資訊獲取到決策輔助 (Impact)
透過介接臺北市政府衛生局的即時稽查資料，市民可以：
- **即時掌握周遭風險**：透過地圖組件，快速查看行政區內商家的稽查合格率與改善狀況。
- **數據驅動決策**：在用餐或採買前，根據視覺化指標判斷商家的信譽，降低暴露於食安風險的機率。
- **強化監督力道**：公開透明的數據能促使商家更積極地改善環境衛生，形成正向的食安循環。

### 0.3 開發功能與程式碼關聯 (Developed Features & Code)
本專案實作了以下核心功能，對應的程式碼如下：
- **數據自動化導入**：透過 `import_food_safety.py` 介接臺北資料大平台 API，自動解析並注入空間座標數據。
- **基礎稽查地圖與趨勢**：由 `app/controllers/initFoodSafety.go` 實作，包含合格率趨勢 (`food_safety_trend`) 與風險評分模型 (`food_safety_risk`)。
- **跨城市比較分析**：由 `app/controllers/initFoodSafetyDualCity.go` 負責，提供雙北案件量與裁罰金額的對比分析。
- **1999 檢舉熱點分析**：由 `app/controllers/initFoodSafetyHeatmap.go` 實作，利用熱度圖標示市民檢舉的空間聚集點，輔助精準稽查。

---

## 1. 概述 (Overview)
食安組件旨在提供台北市（及跨城市）的食品安全相關數據可視化，包含稽查地圖、合格率趨勢、風險權重分析以及檢舉熱點分析。該系統由後端 Gin 框架驅動，並透過 PostgreSQL/PostGIS 進行空間數據處理。

## 2. 數據庫結構 (Database Schema)

### 2.1 `food_safety_inspections` (食安稽查明細)
存儲個別餐廳或商家的稽查結果。
| 欄位名稱 | 類型 | 描述 |
| :--- | :--- | :--- |
| `id` | SERIAL | 主鍵 |
| `name` | VARCHAR | 店家名稱 |
| `address` | TEXT | 地址 |
| `lat`, `lng` | DOUBLE | 經緯度座標 |
| `category` | VARCHAR | 食品或店家類別 |
| `test_item` | VARCHAR | 檢驗項目 |
| `status` | VARCHAR | 狀態 (PASS, FAIL, PENDING) |
| `inspection_date` | DATE | 稽查日期 |

### 2.2 `fact_food_safety_cases` (跨區食安案件)
用於雙北跨城市比較。
| 欄位名稱 | 類型 | 描述 |
| :--- | :--- | :--- |
| `city` | VARCHAR | 城市名稱 (臺北市, 新北市) |
| `case_date` | DATE | 案件日期 |
| `case_type` | VARCHAR | 違規類型 |
| `business_type` | VARCHAR | 業別 (如：超市、夜市) |
| `penalty_amount` | INT | 裁罰金額 |

### 2.3 `food_safety_complaints` (1999 檢舉熱點)
存儲 1999 市民熱線相關食安檢舉。
| 欄位名稱 | 類型 | 描述 |
| :--- | :--- | :--- |
| `case_type` | VARCHAR | 檢舉類型 (如：環境髒亂、疑似中毒) |
| `district` | VARCHAR | 行政區 |
| `lat`, `lng` | DOUBLE | 座標 |

## 3. 組件詳細說明 (Component Details)

### 3.1 臺北食安稽查地圖 (`food_safety`)
- **類型**: GeoJSON Map
- **功能**: 在地圖上標示稽查點，綠色代表合格 (PASS)，紅色代表不合格 (FAIL)，黃色代表處理中 (PENDING)。
- **數據源**: 臺北市政府衛生局

### 3.2 抽驗合格率趨勢 (`food_safety_trend`)
- **類型**: `TimelineSeparateChart`
- **功能**: 顯示每月食品抽驗的合格率走勢（百分比）。
- **邏輯**: `sum(PASS) / total_count * 100`

### 3.3 食品類別風險權重 (`food_safety_risk`)
- **類型**: `RadarChart` (或 3D 圖表)
- **功能**: 綜合「違規率」與「類別基礎嚴重度」計算出的風險評分。
- **風險分級**:
  - 高風險: 生食海鮮、肉類乳製
  - 中高風險: 生鮮蔬果
  - 中風險: 飲冰品豆類
  - 低風險: 乾貨烘焙

### 3.4 雙北案件數比較 (`dual_city_cases`)
- **類型**: `ColumnChart`
- **功能**: 比較臺北市與新北市在不同食安違規類型（如：標示不實、過期食品）的案件數量。

### 3.5 食安檢舉熱點分析 (`food_safety_complaint_heatmap`)
- **類型**: `HeatmapChart` (Mapbox Heatmap Layer)
- **功能**: 分析檢舉案件的空間聚集性，協助精準稽查。

## 4. 初始化 API (Initialization APIs)
系統提供以下端點進行數據表初始化與組件註冊：
- `GET /api/v1/init-food-safety`: 初始化台北食安基礎數據與儀表板。
- `GET /api/v1/init-food-safety-dual-city`: 初始化雙北比較數據。
- `GET /api/v1/init-food-safety-heatmap`: 初始化檢舉熱點數據並更新至食安儀表板。

## 5. 儀表板配置 (Dashboard Configuration)
- **儀表板 ID**: `food_safety`
- **名稱**: 食安地圖與趨勢
- **圖示**: `restaurant`
- **包含組件**:
  1. 臺北食安稽查地圖
  2. 抽驗合格率趨勢
  3. 食品類別風險權重
  4. 食安檢舉熱點分析
  5. 檢舉案件類型統計
