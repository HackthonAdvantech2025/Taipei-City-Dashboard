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
