package controllers

import (
	"TaipeiCityDashboardBE/app/models"
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
)

func InitFoodSafetyHeatmap(c *gin.Context) {
	// 1. Create table for complaints
	models.DBDashboard.Exec("DROP TABLE IF EXISTS food_safety_complaints")
	
	createTableSQL := `
	CREATE TABLE food_safety_complaints (
		id SERIAL PRIMARY KEY,
		case_type VARCHAR(100), -- 過期食品, 環境髒亂, 疑似中毒, 標示不實, 異物
		district VARCHAR(50),
		lat DOUBLE PRECISION,
		lng DOUBLE PRECISION,
		incident_date DATE,
		description TEXT
	);`
	
	if err := models.DBDashboard.Exec(createTableSQL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create table: " + err.Error()})
		return
	}

	// 2. Insert mock 1999 complaint data (focused on certain clusters)
	insertDataSQL := `
	INSERT INTO food_safety_complaints (case_type, district, lat, lng, incident_date, description) 
	VALUES 
		-- Wanhua Cluster (Ximending area)
		('環境髒亂', '萬華區', 25.043, 121.508, '2024-04-05', '西門町某小吃店環境不潔'),
		('環境髒亂', '萬華區', 25.044, 121.507, '2024-04-06', '餐廳廚房有蟑螂'),
		('過期食品', '萬華區', 25.042, 121.509, '2024-04-08', '超市販售過期鮮奶'),
		('環境髒亂', '萬華區', 25.0435, 121.5085, '2024-04-10', '巷弄攤販衛生堪慮'),
		
		-- Da'an Cluster (East District)
		('疑似中毒', '大安區', 25.041, 121.545, '2024-04-12', '多人在燒肉店用餐後不適'),
		('疑似中毒', '大安區', 25.0415, 121.5455, '2024-04-12', '同餐廳疑似中毒報案2'),
		('標示不實', '大安區', 25.033, 121.529, '2024-04-15', '進口水果標示產地不符'),
		('異物', '大安區', 25.025, 121.530, '2024-04-18', '手搖飲內發現塑膠片'),
		
		-- Shilin Cluster (Night Market)
		('過期食品', '士林區', 25.088, 121.524, '2024-04-20', '士林夜市攤位食材不新鮮'),
		('環境髒亂', '士林區', 25.089, 121.525, '2024-04-21', '夜市後方水源污染'),
		('過期食品', '士林區', 25.087, 121.523, '2024-04-22', '零售店販售過期乾貨'),
		
		-- Neihu
		('異物', '內湖區', 25.082, 121.570, '2024-04-25', '便當內有鐵絲'),
		('標示不實', '內湖區', 25.068, 121.574, '2024-04-26', '麵包店宣稱無添加卻含有色素');
	`
    models.DBDashboard.Exec(insertDataSQL)

    // 3. Register components
    var heatmapCompID, barCompID int64
    
    // 3a. Heatmap Component (Represented as a Map component with Heatmap layer)
    models.DBManager.Raw("INSERT INTO components (index, name) VALUES ('food_safety_complaint_heatmap', '食安檢舉熱點分析') ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name RETURNING id").Scan(&heatmapCompID)
    
    // 3b. Bar Chart Component
    models.DBManager.Raw("INSERT INTO components (index, name) VALUES ('food_safety_complaint_types', '檢舉案件類型統計') ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name RETURNING id").Scan(&barCompID)

    // 4. Register chart configs
    // Heatmap (Map) Config - DistrictChart as chart fallback, HeatmapChart needs map layer
    models.DBManager.Exec(`INSERT INTO component_charts (index, color, types, unit) 
                 VALUES ('food_safety_complaint_heatmap', ARRAY['#FFEB3B','#FF9800','#F44336'], ARRAY['HeatmapChart', 'DistrictChart'], '件')
                 ON CONFLICT (index) DO UPDATE SET color = EXCLUDED.color, types = EXCLUDED.types, unit = EXCLUDED.unit`)

    // Bar Config
    models.DBManager.Exec(`INSERT INTO component_charts (index, color, types, unit) 
                 VALUES ('food_safety_complaint_types', ARRAY['#F44336'], ARRAY['BarChart'], '件')
                 ON CONFLICT (index) DO UPDATE SET color = EXCLUDED.color, types = EXCLUDED.types, unit = EXCLUDED.unit`)

    // 5. Map Config for the heatmap layer
    var mapID int
    models.DBManager.Exec("DELETE FROM component_maps WHERE index = 'food_safety_complaint_heatmap'")
    // Using 'heatmap' type for Mapbox
    mapSQL := `INSERT INTO component_maps (index, title, type, source, paint) 
               VALUES ('food_safety_complaint_heatmap', '檢舉熱區', 'heatmap', 'api', '{"heatmap-weight": 1, "heatmap-intensity": 1, "heatmap-radius": 20, "heatmap-opacity": 0.7, "heatmap-color": ["interpolate", ["linear"], ["heatmap-density"], 0, "rgba(33,102,172,0)", 0.2, "rgb(103,169,207)", 0.4, "rgb(209,229,240)", 0.6, "rgb(253,219,199)", 0.8, "rgb(239,138,98)", 1, "rgb(178,24,43)"]}')
               RETURNING id`
    models.DBManager.Raw(mapSQL).Scan(&mapID)

    // 6. Register Queries
    
    // Heatmap GeoJSON Query
    heatmapQuery := `SELECT jsonb_build_object('type', 'FeatureCollection', 'features', jsonb_agg(features.feature))
                 FROM (SELECT jsonb_build_object('type', 'Feature', 'geometry', ST_AsGeoJSON(ST_MakePoint(lng, lat))::jsonb,
                 'properties', jsonb_build_object('type', case_type, 'district', district)) AS feature
                 FROM food_safety_complaints) features`
    
    models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'food_safety_complaint_heatmap' AND city = 'taipei'")
    qcMapSQL := fmt.Sprintf(`INSERT INTO query_charts (index, city, query_type, query_chart, map_config_ids, source, short_desc, created_at, updated_at, time_from, time_to) 
              VALUES ('food_safety_complaint_heatmap', 'taipei', 'geojson', ?, ARRAY[%d], '1999 市民熱線', '分析食安檢舉熱點，協助精準稽查。', NOW(), NOW(), 'max', 'now')`, mapID)
    models.DBManager.Exec(qcMapSQL, heatmapQuery)

    // Bar Chart Query
    barQuery := `SELECT 
                    case_type as x_axis,
                    '檢舉件數' as y_axis,
                    count(*) as data
                FROM food_safety_complaints
                GROUP BY 1, 2
                ORDER BY 3 DESC`
    
    models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'food_safety_complaint_types' AND city = 'taipei'")
    models.DBManager.Exec(`INSERT INTO query_charts (index, city, query_type, query_chart, source, short_desc, created_at, updated_at, time_from, time_to) 
              VALUES ('food_safety_complaint_types', 'taipei', 'two_d', ?, '1999 市民熱線', '統計最常被檢舉的食安問題類型。', NOW(), NOW(), 'max', 'now')`, barQuery)

    // 7. Create a dedicated dashboard for heatmap analysis
    var dashID int
    dashSQL := `INSERT INTO dashboards (index, name, components, icon, created_at, updated_at) 
                VALUES ('food_safety_heatmap', '食安檢舉熱點分析', ARRAY[?, ?]::integer[], 'location_on', NOW(), NOW())
                ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name, components = EXCLUDED.components, icon = EXCLUDED.icon, updated_at = NOW()
                RETURNING id`
    models.DBManager.Raw(dashSQL, heatmapCompID, barCompID).Scan(&dashID)

    // Also append to food_safety dashboard using PostgreSQL array functions
    models.DBManager.Exec(`
        UPDATE dashboards SET 
            components = ARRAY(SELECT DISTINCT unnest(components || ARRAY[?, ?]::integer[]))
        WHERE index = 'food_safety'`, heatmapCompID, barCompID)

    // Assign heatmap dashboard to taipei group (group_id=2)
    models.DBManager.Exec("DELETE FROM dashboard_groups WHERE dashboard_id = ?", dashID)
    models.DBManager.Exec("INSERT INTO dashboard_groups (dashboard_id, group_id) VALUES (?, 2)", dashID)

	c.JSON(http.StatusOK, gin.H{
        "message": "Incident heatmap initialization successful", 
        "heatmap_id": heatmapCompID, 
        "bar_id": barCompID,
    })
}
