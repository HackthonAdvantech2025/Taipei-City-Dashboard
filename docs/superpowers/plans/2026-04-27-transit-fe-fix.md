# Transit FE Fix — PLAN Alignment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修正 feature/transit-fe 四個與 PLAN 不符的偏差：組件路徑、DashboardView 嵌入、transitStore axios 實例、mock 資料格式。

**Architecture:** 將 TransitMap/ArrivalBoard 移至 `dashboardComponent/components/`，改成 props/emit 模式；transitStore 改用專案 `http` 實例；DashboardView 底部嵌入 transit section；mock 資料對齊 PLAN 格式。

**Tech Stack:** Vue 3 (`<script setup>`), Pinia, Mapbox GL JS, 專案 `http`（axios wrapper）

**Worktree:** `/home/nelson/transit-fe/Taipei-City-Dashboard-FE`

---

## 檔案結構

| 動作 | 路徑 |
|---|---|
| 修改 | `src/assets/mock/transit-stations.json` |
| 修改 | `src/assets/mock/transit-arrivals.json` |
| 修改 | `src/store/transitStore.js` |
| 新建 | `src/dashboardComponent/components/TransitMap.vue` |
| 新建 | `src/dashboardComponent/components/ArrivalBoard.vue` |
| 修改 | `src/views/DashboardView.vue` |
| 修改 | `src/views/TransitView.vue` (更新 import 路徑) |
| 刪除 | `src/components/transit/TransitMap.vue` |
| 刪除 | `src/components/transit/ArrivalBoard.vue` |

---

## Task 1：修正 mock 資料（stations + arrivals）

**Files:**
- Modify: `src/assets/mock/transit-stations.json`
- Modify: `src/assets/mock/transit-arrivals.json`

- [ ] **Step 1：替換 transit-stations.json**

完整替換為 PLAN 指定的真實站名，格式為 flat array（store 直接賦值）：

```json
[
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
```

- [ ] **Step 2：替換 transit-arrivals.json**

格式改為扁平 `{"data":[...]}` 包裝，`estimate_time` 單位秒：

```json
{
  "data": [
    { "station_id": "YL01", "station_name_zh": "新北產業園區", "direction": 0, "destination_zh": "迴龍", "estimate_time": 180, "updated_at": "2026-04-27T10:42:00+08:00" },
    { "station_id": "YL02", "station_name_zh": "幸福", "direction": 0, "destination_zh": "迴龍", "estimate_time": 420, "updated_at": "2026-04-27T10:42:00+08:00" },
    { "station_id": "YL01", "station_name_zh": "新北產業園區", "direction": 1, "destination_zh": "新埔民生", "estimate_time": 300, "updated_at": "2026-04-27T10:42:00+08:00" },
    { "station_id": "A01", "station_name_zh": "台北車站", "direction": 0, "destination_zh": "桃園機場", "estimate_time": 720, "updated_at": "2026-04-27T10:42:00+08:00" },
    { "station_id": "A01", "station_name_zh": "台北車站", "direction": 1, "destination_zh": "台北車站", "estimate_time": 480, "updated_at": "2026-04-27T10:42:00+08:00" }
  ]
}
```

- [ ] **Step 3：Commit**

```bash
cd /home/nelson/transit-fe
git add Taipei-City-Dashboard-FE/src/assets/mock/
git commit -m "fix(fe): 對齊 PLAN mock 資料格式與真實站名"
```

---

## Task 2：修正 transitStore.js

**Files:**
- Modify: `src/store/transitStore.js`

- [ ] **Step 1：完整替換 transitStore.js**

```js
// Developed by Taipei Urban Intelligence Center 2023-2024

/* transitStore */
/*
The transitStore handles transit stations and arrivals data.
*/

import { defineStore } from "pinia";
import http from "../router/axios.js";
import stationsMock from "../assets/mock/transit-stations.json";
import arrivalsMock from "../assets/mock/transit-arrivals.json";

const USE_MOCK = false;

export const useTransitStore = defineStore("transit", {
	state: () => ({
		stations: [],
		arrivals: [],
		selectedStation: null,
		activeSystem: "ntmetro",
		loading: false,
		error: null,
	}),
	getters: {
		filteredStations: (state) => {
			return state.stations.filter((s) => s.system === state.activeSystem);
		},
	},
	actions: {
		async fetchStations() {
			this.loading = true;
			this.error = null;
			try {
				if (USE_MOCK) {
					this.stations = stationsMock;
				} else {
					const response = await http.get("/transit/stations", {
						params: { system: "ntmetro,tymc" },
					});
					this.stations = response.data.data;
				}
			} catch (err) {
				this.error = err.message || "Failed to fetch stations";
			} finally {
				this.loading = false;
			}
		},
		async fetchArrivals(system, stationId) {
			this.loading = true;
			this.error = null;
			try {
				if (USE_MOCK) {
					const all = arrivalsMock.data;
					this.arrivals = stationId
						? all.filter((a) => a.station_id === stationId)
						: all.filter((a) => {
							const st = this.stations.find((s) => s.station_id === a.station_id);
							return st?.system === system;
						});
				} else {
					const params = { system };
					if (stationId) params.station_id = stationId;
					const response = await http.get("/transit/arrivals", { params });
					this.arrivals = response.data.data;
				}
			} catch (err) {
				this.error = err.message || "Failed to fetch arrivals";
			} finally {
				this.loading = false;
			}
		},
		setSelectedStation(station) {
			this.selectedStation = station;
			if (station) {
				this.fetchArrivals(station.system, station.station_id);
			} else {
				this.arrivals = [];
			}
		},
		setActiveSystem(system) {
			this.activeSystem = system;
		},
	},
});
```

**注意：** `http` 的 `baseURL` 是 `VITE_API_URL`（生產環境為 `/api/v1`），所以路徑寫 `/transit/stations`（無 `/api/v1` 前綴）。Dev 環境 proxy 設定為 `/api/dev` → `dashboard-be:8080/api/v1`，因此 dev 時需確認 `.env.development` 有 `VITE_API_URL=/api/dev`。

- [ ] **Step 2：Commit**

```bash
cd /home/nelson/transit-fe
git add Taipei-City-Dashboard-FE/src/store/transitStore.js
git commit -m "fix(fe): transitStore 改用專案 http 實例，對齊 mock 資料存取格式"
```

---

## Task 3：建立新 TransitMap.vue（dashboardComponent/components/）

**Files:**
- Create: `src/dashboardComponent/components/TransitMap.vue`

- [ ] **Step 1：確認目標資料夾存在**

```bash
ls /home/nelson/transit-fe/Taipei-City-Dashboard-FE/src/dashboardComponent/components/ | head -5
```

預期：列出現有組件（如 `ColumnChart.vue` 等）。

- [ ] **Step 2：建立 TransitMap.vue**

```vue
<script setup>
import { ref, onMounted, onUnmounted, watch, computed } from 'vue';
import mapboxGl from 'mapbox-gl';
import 'mapbox-gl/dist/mapbox-gl.css';
import { useTransitStore } from '../../store/transitStore';

const props = defineProps({
	systems: { type: Array, default: () => ['ntmetro', 'tymc'] },
});
const emit = defineEmits(['station-click']);

const transitStore = useTransitStore();
const mapContainer = ref(null);
const map = ref(null);
const markers = ref([]);

const activeSystem = computed(() => transitStore.activeSystem);
const filteredStations = computed(() => transitStore.filteredStations);

const MAPBOX_TOKEN = import.meta.env.VITE_MAPBOXTOKEN;
const MAPBOX_STYLE = import.meta.env.VITE_MAPBOXTILE || 'mapbox://styles/mapbox/dark-v10';

const SYSTEM_CONFIG = {
	ntmetro: { color: '#eab308', label: '環狀線' },
	tymc: { color: '#7c3aed', label: '機捷' },
};

const initMap = () => {
	mapboxGl.accessToken = MAPBOX_TOKEN;
	map.value = new mapboxGl.Map({
		container: mapContainer.value,
		style: MAPBOX_STYLE,
		center: [121.46, 25.02],
		zoom: 10,
	});
	map.value.on('load', () => {
		renderMarkers();
	});
};

const clearMarkers = () => {
	markers.value.forEach((marker) => marker.remove());
	markers.value = [];
};

const renderMarkers = () => {
	if (!map.value) return;
	clearMarkers();
	const stations = filteredStations.value;
	if (stations.length === 0) return;
	const bounds = new mapboxGl.LngLatBounds();
	stations.forEach((station) => {
		const {lng, lat} = station;
		if (!lng || !lat) return;
		const el = document.createElement('div');
		el.style.width = '12px';
		el.style.height = '12px';
		el.style.borderRadius = '50%';
		el.style.backgroundColor = SYSTEM_CONFIG[station.system]?.color || '#ffffff';
		el.style.cursor = 'pointer';
		el.style.border = '2px solid white';
		el.addEventListener('click', () => {
			emit('station-click', station.station_id, station.system);
		});
		const marker = new mapboxGl.Marker({ element: el })
			.setLngLat([lng, lat])
			.addTo(map.value);
		markers.value.push(marker);
		bounds.extend([lng, lat]);
	});
	if (!bounds.isEmpty()) {
		map.value.fitBounds(bounds, { padding: 50, maxZoom: 13 });
	}
};

watch(activeSystem, () => {
	renderMarkers();
});

onMounted(() => {
	initMap();
});

onUnmounted(() => {
	if (map.value) {
		map.value.remove();
	}
});

const handleToggleSystem = (system) => {
	transitStore.setActiveSystem(system);
};
</script>

<template>
  <div class="transit-map-wrapper">
    <div class="system-toggles">
      <button
        v-for="sys in props.systems"
        :key="sys"
        :class="{ active: activeSystem === sys }"
        :style="{ '--color': SYSTEM_CONFIG[sys]?.color }"
        @click="handleToggleSystem(sys)"
      >
        {{ SYSTEM_CONFIG[sys]?.label || sys }}
      </button>
    </div>
    <div
      ref="mapContainer"
      class="map-container"
    />
  </div>
</template>

<style scoped>
.transit-map-wrapper {
  width: 100%;
}
.system-toggles {
  display: flex;
  gap: 10px;
  padding: 8px 12px;
  background: #1e293b;
}
button {
  padding: 4px 12px;
  border-radius: 20px;
  border: 2px solid var(--color, #ccc);
  background: transparent;
  color: var(--color, #ccc);
  cursor: pointer;
  font-size: 12px;
  transition: background 0.2s;
}
button.active {
  background: var(--color, #ccc);
  color: #0f172a;
  font-weight: bold;
}
.map-container {
  height: 400px;
  width: 100%;
  border-radius: 0 0 8px 8px;
  overflow: hidden;
}
</style>
```

- [ ] **Step 3：Commit**

```bash
cd /home/nelson/transit-fe
git add Taipei-City-Dashboard-FE/src/dashboardComponent/components/TransitMap.vue
git commit -m "fix(fe): 新增 TransitMap.vue 至 dashboardComponent/components（emit pattern）"
```

---

## Task 4：建立新 ArrivalBoard.vue（dashboardComponent/components/）

**Files:**
- Create: `src/dashboardComponent/components/ArrivalBoard.vue`

- [ ] **Step 1：建立 ArrivalBoard.vue**

```vue
<template>
  <div class="arrival-board">
    <div
      v-if="!props.stationId"
      class="placeholder"
    >
      點選地圖上的站點查看到站時間
    </div>

    <div
      v-else
      class="board-content"
    >
      <header class="board-header">
        <h3>{{ currentStationName }}</h3>
        <span class="header-label">到站資訊</span>
        <div
          v-if="transitStore.loading"
          class="loading-indicator"
        >
          <span class="dot" />
          <span class="dot" />
          <span class="dot" />
        </div>
      </header>

      <div
        v-if="transitStore.arrivals.length === 0 && !transitStore.loading"
        class="no-data"
      >
        暫無到站資訊
      </div>

      <div
        v-else
        class="arrival-list"
      >
        <div
          v-for="(arrival, index) in transitStore.arrivals"
          :key="`${arrival.station_id}-${arrival.direction}-${index}`"
          class="arrival-item"
        >
          <div class="arrival-main-info">
            <div class="arrival-route">
              <span
                class="direction-tag"
                :class="{ 'return': arrival.direction === 1 }"
              >
                {{ arrival.direction === 0 ? '去程' : '返程' }}
              </span>
            </div>
            <div class="arrival-destination">
              往 {{ arrival.destination_zh }}
            </div>
            <div class="arrival-time">
              {{ formatTime(arrival.estimate_time) }}
            </div>
          </div>
          <div class="progress-container">
            <div
              class="progress-bar"
              :style="{ width: Math.min((arrival.estimate_time / 300) * 100, 100) + '%' }"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, watch } from 'vue'
import { useTransitStore } from '../../store/transitStore'

const props = defineProps({
	system: { type: String, default: 'ntmetro' },
	stationId: { type: String, default: null },
})

const transitStore = useTransitStore()
let refreshInterval = null

const currentStationName = computed(() => {
	const st = transitStore.stations.find((s) => s.station_id === props.stationId)
	return st?.station_name_zh || props.stationId || ''
})

const formatTime = (seconds) => {
	if (seconds <= 0) return '進站中'
	const mins = Math.floor(seconds / 60)
	const secs = seconds % 60
	if (mins > 0) return `${mins}分${secs.toString().padStart(2, '0')}秒`
	return `${secs}秒`
}

const doFetch = () => {
	if (props.stationId) {
		transitStore.fetchArrivals(props.system, props.stationId)
	}
}

const startRefresh = () => {
	stopRefresh()
	refreshInterval = setInterval(doFetch, 30000)
}

const stopRefresh = () => {
	if (refreshInterval) {
		clearInterval(refreshInterval)
		refreshInterval = null
	}
}

watch(() => [props.system, props.stationId], ([, newId]) => {
	if (newId) {
		doFetch()
		startRefresh()
	} else {
		transitStore.arrivals = []
		stopRefresh()
	}
}, { immediate: true })

onMounted(() => {
	if (props.stationId) startRefresh()
})

onUnmounted(() => {
	stopRefresh()
})
</script>

<style scoped>
.arrival-board {
  background-color: #1e2026;
  color: #ffffff;
  padding: 1.25rem;
  border-radius: 8px;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.placeholder, .no-data {
  display: flex;
  justify-content: center;
  align-items: center;
  flex-grow: 1;
  color: #888;
  font-size: 0.95rem;
  text-align: center;
  padding: 2rem;
}
.board-content {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.board-header {
  display: flex;
  align-items: baseline;
  gap: 0.75rem;
  margin-bottom: 1.25rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid #333;
}
.board-header h3 {
  margin: 0;
  font-size: 1.35rem;
  color: #00e5ff;
  font-weight: 600;
}
.header-label { font-size: 0.85rem; color: #aaa; }
.loading-indicator { display: flex; gap: 4px; margin-left: auto; }
.dot {
  width: 6px; height: 6px;
  background-color: #00e5ff;
  border-radius: 50%;
  animation: pulse 1.5s infinite ease-in-out;
}
.dot:nth-child(2) { animation-delay: 0.2s; }
.dot:nth-child(3) { animation-delay: 0.4s; }
@keyframes pulse {
  0%, 100% { opacity: 0.3; transform: scale(0.8); }
  50% { opacity: 1; transform: scale(1.1); }
}
.arrival-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  overflow-y: auto;
}
.arrival-item {
  background-color: #2a2d36;
  padding: 1rem;
  border-radius: 6px;
}
.arrival-main-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}
.direction-tag {
  background-color: #00c853;
  color: white;
  padding: 1px 6px;
  border-radius: 3px;
  font-size: 0.75rem;
}
.direction-tag.return { background-color: #ff6d00; }
.arrival-destination { flex-grow: 1; margin: 0 1rem; font-weight: 500; }
.arrival-time { color: #ffca28; font-weight: 700; min-width: 80px; text-align: right; }
.progress-container { height: 4px; background-color: #111; border-radius: 2px; overflow: hidden; }
.progress-bar { height: 100%; background: linear-gradient(90deg, #00e5ff, #00b0ff); }
</style>
```

- [ ] **Step 2：Commit**

```bash
cd /home/nelson/transit-fe
git add Taipei-City-Dashboard-FE/src/dashboardComponent/components/ArrivalBoard.vue
git commit -m "fix(fe): 新增 ArrivalBoard.vue 至 dashboardComponent/components（props pattern）"
```

---

## Task 5：修改 DashboardView.vue（嵌入 transit section）

**Files:**
- Modify: `src/views/DashboardView.vue`

- [ ] **Step 1：在 `<script setup>` 區塊加入 imports 和 handlers**

在 `DashboardView.vue` 的 `<script setup>` 區塊，找到最後一個 import 行（`import ReportIssue...`）後加入：

```js
import { ref } from "vue";
import TransitMap from "../dashboardComponent/components/TransitMap.vue";
import ArrivalBoard from "../dashboardComponent/components/ArrivalBoard.vue";
import { useTransitStore } from "../store/transitStore";

const transitStore = useTransitStore();
const arrivalSystem = ref("ntmetro");
const arrivalStationId = ref(null);

function onTransitStationClick(stationId, system) {
	arrivalSystem.value = system;
	arrivalStationId.value = stationId;
}
```

並在最後一個 function（`handleMoreInfo`）結束的 `}` 後、`</script>` 前，加入：

```js
// 初始化站點資料（只需一次）
transitStore.fetchStations();
```

- [ ] **Step 2：在 `<template>` 最後加入 transit section**

找到 `</template>` 前的最後一個閉合 `</div>`，在它**之後、`</template>` 之前**加入：

```vue
<!-- Transit Map PoC Section -->
<div class="transit-poc-section">
  <h3 class="transit-poc-title">
    跨運具整合地圖
  </h3>
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

- [ ] **Step 3：在 `<style>` 區塊加入樣式**

在現有 `<style scoped lang="scss">` 的 `@keyframes spin { ... }` 後加入：

```scss
.transit-poc-section {
	padding: 24px var(--font-m);
	background: #0a0e1a;
	grid-column: 1 / -1;
}
.transit-poc-title {
	color: #94a3b8;
	font-size: 14px;
	text-transform: uppercase;
	letter-spacing: 1px;
	margin-bottom: 16px;
	font-weight: 600;
}
.transit-poc-grid {
	display: grid;
	grid-template-columns: 2fr 1fr;
	gap: 16px;
	min-height: 360px;

	@media (max-width: 720px) {
		grid-template-columns: 1fr;
	}
}
.transit-poc-map-card,
.transit-poc-arrival-card {
	background: #0f172a;
	border-radius: 8px;
	overflow: hidden;
	border: 1px solid #1e293b;
}
```

- [ ] **Step 4：Commit**

```bash
cd /home/nelson/transit-fe
git add Taipei-City-Dashboard-FE/src/views/DashboardView.vue
git commit -m "fix(fe): 在 DashboardView 嵌入 transit section（符合 PLAN Task 12）"
```

---

## Task 6：更新 TransitView.vue imports + 刪除舊組件

**Files:**
- Modify: `src/views/TransitView.vue`
- Delete: `src/components/transit/TransitMap.vue`
- Delete: `src/components/transit/ArrivalBoard.vue`

- [ ] **Step 1：更新 TransitView.vue 的 import 路徑**

在 `src/views/TransitView.vue` 的 `<script setup>` 中，將：

```js
import TransitMap from '../components/transit/TransitMap.vue'
import ArrivalBoard from '../components/transit/ArrivalBoard.vue'
```

改為：

```js
import TransitMap from '../dashboardComponent/components/TransitMap.vue'
import ArrivalBoard from '../dashboardComponent/components/ArrivalBoard.vue'
```

- [ ] **Step 2：刪除舊組件目錄**

```bash
rm /home/nelson/transit-fe/Taipei-City-Dashboard-FE/src/components/transit/TransitMap.vue
rm /home/nelson/transit-fe/Taipei-City-Dashboard-FE/src/components/transit/ArrivalBoard.vue
rmdir /home/nelson/transit-fe/Taipei-City-Dashboard-FE/src/components/transit/
```

- [ ] **Step 3：Commit**

```bash
cd /home/nelson/transit-fe
git add -A Taipei-City-Dashboard-FE/src/components/transit/
git add Taipei-City-Dashboard-FE/src/views/TransitView.vue
git commit -m "fix(fe): 移除舊組件路徑，更新 TransitView imports"
```

---

## Task 7：Build 驗證

**Files:** 無新增

- [ ] **Step 1：執行 build**

```bash
cd /home/nelson/transit-fe/Taipei-City-Dashboard-FE
npm run build 2>&1 | tail -20
```

預期：`✓ built in XX.XXs`，無 error，`TransitView` 和 `DashboardView` 均出現在 chunk 清單中。

- [ ] **Step 2：確認所有 transit 組件路徑正確**

```bash
grep -r "components/transit" /home/nelson/transit-fe/Taipei-City-Dashboard-FE/src/ || echo "無殘留舊路徑"
```

預期：輸出「無殘留舊路徑」。

- [ ] **Step 3：確認 DashboardView 有 transit section**

```bash
grep -n "transit-poc-section\|onTransitStationClick\|TransitMap\|ArrivalBoard" \
  /home/nelson/transit-fe/Taipei-City-Dashboard-FE/src/views/DashboardView.vue
```

預期：至少 6 行匹配。

---

## Self-Review

**Spec coverage check:**
- [x] 組件移至 dashboardComponent/components/ → Task 3, 4, 6
- [x] DashboardView 嵌入 transit section → Task 5
- [x] transitStore 使用 http → Task 2
- [x] Mock 資料格式對齊 → Task 1
- [x] 舊組件刪除 → Task 6
- [x] Build 驗證 → Task 7

**Placeholder scan:** 無 TBD / TODO。

**Type consistency:**
- `transitStore.fetchArrivals(system, stationId)` — Task 2 定義，Task 4 呼叫 ✅
- `transitStore.fetchStations()` — Task 2 定義，Task 5 呼叫 ✅
- `emit('station-click', stationId, system)` — Task 3 定義，Task 5 接收 `onTransitStationClick(stationId, system)` ✅
- `props.stationId` / `props.system` — Task 4 定義，Task 5 傳入 `:station-id` / `:system` ✅
