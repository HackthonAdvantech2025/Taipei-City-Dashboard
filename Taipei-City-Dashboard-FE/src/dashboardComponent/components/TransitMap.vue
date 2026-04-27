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
