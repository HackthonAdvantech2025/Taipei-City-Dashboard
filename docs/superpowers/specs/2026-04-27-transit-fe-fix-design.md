# Transit FE Fix Design — PLAN Alignment

**Date:** 2026-04-27  
**Scope:** `feature/transit-fe` worktree  
**Goal:** Correct four deviations found during PLAN cross-check.

---

## Deviations Being Fixed

| # | Item | Was | Should Be |
|---|---|---|---|
| 1 | Component path | `src/components/transit/` | `src/dashboardComponent/components/` |
| 2 | Integration point | Standalone `/transit` route only | Hardcoded section in `DashboardView.vue` |
| 3 | transitStore axios | plain `axios` | project `http` from `../router/axios.js` |
| 4 | Mock data | Wrong station names + system-keyed arrivals | PLAN station names + flat `{"data":[...]}` arrivals |

---

## Architecture

### Component Event Flow (following PLAN Task 12)

```
DashboardView.vue
├── <TransitMap @station-click="onTransitStationClick" />
│     emits: station-click(stationId, system)
└── <ArrivalBoard :system="arrivalSystem" :station-id="arrivalStationId" />
      watches props → calls transitStore.fetchArrivals(system, stationId)
```

DashboardView owns `arrivalSystem` and `arrivalStationId` refs, wires the two components together.

### transitStore.js Changes

- Replace `import axios from 'axios'` → `import http from '../router/axios.js'`
- API calls: `http.get('/transit/stations', ...)` (baseURL from VITE_API_URL handles prefix)
- Mock stations: flat array format, 10 stations from PLAN (real names: 新北產業園區, 幸福, 頭前庄…)
- Mock arrivals: `{"data": [...]}` flat format, `estimate_time` in seconds

### TransitMap.vue (rewrite target)

- Location: `src/dashboardComponent/components/TransitMap.vue`
- Accept `systems` prop (Array, default `['ntmetro','tymc']`)
- Emit `station-click(stationId, system)` on marker click
- Does NOT call store directly for station selection
- Uses Mapbox GL DOM markers (token may be empty in dev, graceful degradation)
- Still calls `transitStore.fetchStations()` on mount for station data

### ArrivalBoard.vue (rewrite target)

- Location: `src/dashboardComponent/components/ArrivalBoard.vue`
- Accept `system` (String) and `stationId` (String|null) props
- Watch props → call `transitStore.fetchArrivals(system, stationId)`
- 30s auto-refresh interval
- Display arrivals from `transitStore.arrivals`

### DashboardView.vue Addition

Append before `</template>` (outside existing v-if blocks):

```vue
<!-- Transit Map PoC Section -->
<div class="transit-poc-section">
  <h3 class="transit-poc-title">跨運具整合地圖</h3>
  <div class="transit-poc-grid">
    <div class="transit-poc-map-card">
      <TransitMap :systems="['ntmetro', 'tymc']" @station-click="onTransitStationClick" />
    </div>
    <div class="transit-poc-arrival-card">
      <ArrivalBoard :system="arrivalSystem" :station-id="arrivalStationId" />
    </div>
  </div>
</div>
```

---

## Files Changed

| File | Action |
|---|---|
| `src/assets/mock/transit-stations.json` | Replace content (PLAN station names) |
| `src/assets/mock/transit-arrivals.json` | Replace content (flat `data` format) |
| `src/store/transitStore.js` | http + mock access pattern |
| `src/dashboardComponent/components/TransitMap.vue` | New (props/emit pattern) |
| `src/dashboardComponent/components/ArrivalBoard.vue` | New (props pattern) |
| `src/views/DashboardView.vue` | Add transit section |
| `src/components/transit/` | Delete old files |
| `src/views/TransitView.vue` | Update imports to new paths |

---

## Out of Scope

- Mapbox GeoJSON layer approach (DOM markers kept for token-agnostic dev)
- `/transit` route removal (kept as bonus Demo entry point)
- DE / BE changes (none needed)
