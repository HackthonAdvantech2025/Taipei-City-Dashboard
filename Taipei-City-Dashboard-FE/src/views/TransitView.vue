<template>
  <div class="transit-view">
    <header class="transit-header">
      <div class="title-group">
        <h1>跨運具整合地圖</h1>
        <p class="subtitle">
          環狀線 · 機捷 即時到站資訊
        </p>
      </div>
    </header>
    <div class="transit-content">
      <div class="map-container">
        <TransitMap />
      </div>
      <div class="board-container">
        <ArrivalBoard />
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import TransitMap from '../dashboardComponent/components/TransitMap.vue'
import ArrivalBoard from '../dashboardComponent/components/ArrivalBoard.vue'
import { useTransitStore } from '../store/transitStore'

const transitStore = useTransitStore()

onMounted(() => {
	transitStore.fetchStations()
})
</script>

<style scoped lang="scss">
.transit-view {
  background-color: #0e1118;
  height: calc(100vh - 127px);
  height: calc(var(--vh, 1vh) * 100 - 127px);
  color: white;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.transit-header {
  padding: 1.25rem 2rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  background-color: rgba(14, 17, 24, 0.8);
  backdrop-filter: blur(8px);
  z-index: 10;
}

.title-group h1 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
  letter-spacing: 0.025em;
}

.subtitle {
  margin: 0.25rem 0 0;
  font-size: 0.875rem;
  color: #94a3b8;
}

.transit-content {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr;
  overflow: hidden;
  
  @media (min-width: 1024px) {
    grid-template-columns: 6fr 4fr;
  }
}

.map-container {
  position: relative;
  height: 100%;
  min-height: 400px;
}

.board-container {
  border-left: 1px solid rgba(255, 255, 255, 0.1);
  overflow-y: auto;
  height: 100%;
  background-color: rgba(15, 23, 42, 0.3);

  @media (max-width: 1023px) {
    border-left: none;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    height: 400px;
  }
}

/* Custom scrollbar for board-container */
.board-container::-webkit-scrollbar {
  width: 6px;
}

.board-container::-webkit-scrollbar-track {
  background: transparent;
}

.board-container::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
}

.board-container::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.2);
}
</style>
