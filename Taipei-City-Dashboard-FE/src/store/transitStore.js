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
