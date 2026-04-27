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
