package controllers

import (
	"TaipeiCityDashboardBE/app/models"
	"net/http"
	"github.com/gin-gonic/gin"
)

func InitSchoolLunch(c *gin.Context) {
	// 1. Create table in dashboard DB
	models.DBDashboard.Exec("DROP TABLE IF EXISTS school_lunch_traceability")
	
	createTableSQL := `
	CREATE TABLE school_lunch_traceability (
		id SERIAL PRIMARY KEY,
		school_name VARCHAR(255),
		district VARCHAR(50),
		cert_type VARCHAR(50), -- Organic, TAP, CAS, QR_Code
		proportion NUMERIC,
		record_date DATE
	);`
	
	if err := models.DBDashboard.Exec(createTableSQL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create table: " + err.Error()})
		return
	}

	// 2. Insert mock data
	insertDataSQL := `
	INSERT INTO school_lunch_traceability (school_name, district, cert_type, proportion, record_date) 
	VALUES 
		-- Da'an District
		('大安國小', '大安區', 'Organic', 25.5, '2026-04-01'),
		('大安國小', '大安區', 'TAP', 30.2, '2026-04-01'),
		('大安國小', '大安區', 'CAS', 20.1, '2026-04-01'),
		('大安國小', '大安區', 'QR_Code', 24.2, '2026-04-01'),
		-- Xinyi District
		('信義國小', '信義區', 'Organic', 22.1, '2026-04-01'),
		('信義國小', '信義區', 'TAP', 35.5, '2026-04-01'),
		('信義國小', '信義區', 'CAS', 18.2, '2026-04-01'),
		('信義國小', '信義區', 'QR_Code', 24.2, '2026-04-01'),
		-- Beitou District
		('北投國小', '北投區', 'Organic', 28.4, '2026-04-01'),
		('北投國小', '北投區', 'TAP', 28.2, '2026-04-01'),
		('北投國小', '北投區', 'CAS', 22.1, '2026-04-01'),
		('北投國小', '北投區', 'QR_Code', 21.3, '2026-04-01'),
		-- Shilin District
		('士林國小', '士林區', 'Organic', 24.5, '2026-04-01'),
		('士林國小', '士林區', 'TAP', 32.1, '2026-04-01'),
		('士林國小', '士林區', 'CAS', 20.4, '2026-04-01'),
		('士林國小', '士林區', 'QR_Code', 23.0, '2026-04-01'),
		-- Neihu District
		('內湖國小', '內湖區', 'Organic', 26.1, '2026-04-01'),
		('內湖國小', '內湖區', 'TAP', 29.8, '2026-04-01'),
		('內湖國小', '內湖區', 'CAS', 21.5, '2026-04-01'),
		('內湖國小', '內湖區', 'QR_Code', 22.6, '2026-04-01');
	`
    // Add more districts for coverage
    models.DBDashboard.Exec(insertDataSQL)
    models.DBDashboard.Exec(`
        INSERT INTO school_lunch_traceability (school_name, district, cert_type, proportion, record_date) 
        VALUES 
            ('中山國小', '中山區', 'Organic', 23.4, '2026-04-01'), ('中山國小', '中山區', 'TAP', 31.2, '2026-04-01'),
            ('萬華國小', '萬華區', 'Organic', 20.5, '2026-04-01'), ('萬華國小', '萬華區', 'TAP', 34.5, '2026-04-01'),
            ('文山國小', '文山區', 'Organic', 27.2, '2026-04-01'), ('文山國小', '文山區', 'TAP', 27.8, '2026-04-01'),
            ('松山國小', '松山區', 'Organic', 25.1, '2026-04-01'), ('松山國小', '松山區', 'TAP', 30.5, '2026-04-01'),
            ('中正國小', '中正區', 'Organic', 24.8, '2026-04-01'), ('中正國小', '中正區', 'TAP', 32.6, '2026-04-01'),
            ('大同國小', '大同區', 'Organic', 21.2, '2026-04-01'), ('大同國小', '大同區', 'TAP', 33.9, '2026-04-01'),
            ('南港國小', '南港區', 'Organic', 26.5, '2026-04-01'), ('南港國小', '南港區', 'TAP', 29.1, '2026-04-01');
    `)

    // 3. Register components
    var distCompID int64
    
    // 3b. District Component
    models.DBManager.Raw("INSERT INTO components (index, name) VALUES ('school_lunch_district', '各區食材溯源達成率') ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name RETURNING id").Scan(&distCompID)

    // 4. Register chart configs
    // District Config
    models.DBManager.Exec(`INSERT INTO component_charts (index, color, types, unit) 
                 VALUES ('school_lunch_district', ARRAY['#E8F5E9','#4CAF50','#1B5E20'], ARRAY['DistrictChart', 'ColumnChart'], '%')
                 ON CONFLICT (index) DO UPDATE SET color = EXCLUDED.color, types = EXCLUDED.types, unit = EXCLUDED.unit`)

    // 5. Register Queries in query_charts
    
    // District Query
    distQuery := `SELECT 
                        district as x_axis,
                        '達成率' as y_axis,
                        round(sum(proportion), 1) as data -- Summing Organic + TAP + CAS + QR_Code
                    FROM school_lunch_traceability
                    GROUP BY 1, 2`
    
    models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'school_lunch_district' AND city = 'taipei'")
    models.DBManager.Exec(`INSERT INTO query_charts (index, city, query_type, query_chart, source, short_desc, created_at, updated_at, time_from, time_to) 
              VALUES ('school_lunch_district', 'taipei', 'two_d', ?, '臺北市政府教育局', '比較臺北市各區學校食材溯源的達成率。', NOW(), NOW(), 'max', 'now')`, distQuery)

    // 6. Append to food_safety dashboard using PostgreSQL array functions
    models.DBManager.Exec(`
        UPDATE dashboards SET 
            components = ARRAY(SELECT DISTINCT unnest(components || ARRAY[?]::integer[]))
        WHERE index = 'food_safety'`, distCompID)

    // Delete the old school_lunch dashboard if it exists
    models.DBManager.Exec("DELETE FROM dashboards WHERE index = 'school_lunch'")

	c.JSON(http.StatusOK, gin.H{
        "message": "School lunch traceability initialization successful and appended to food_safety", 
        "components": []int64{distCompID}, 
    })
}
