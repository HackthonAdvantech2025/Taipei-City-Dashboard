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
