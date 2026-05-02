package controllers

import (
	"TaipeiCityDashboardBE/app/models"
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
)

func InitFoodSafety(c *gin.Context) {
	// 1. Create table and insert data in dashboard DB
	// Drop table first to ensure schema updates are applied
	models.DBDashboard.Exec("DROP TABLE IF EXISTS food_safety_inspections")
	
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS food_safety_inspections (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		address TEXT,
		lat DOUBLE PRECISION,
		lng DOUBLE PRECISION,
		category VARCHAR(100),
		test_item VARCHAR(100),
		status VARCHAR(50),
		inspection_date DATE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`
	
	if err := models.DBDashboard.Exec(createTableSQL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create table: " + err.Error()})
		return
	}
	// Clear existing data
	models.DBDashboard.Exec("TRUNCATE TABLE food_safety_inspections")

	// 2. Insert comprehensive mock data (12 months, 3 categories)
	insertDataSQL := `
	INSERT INTO food_safety_inspections (name, address, lat, lng, category, test_item, status, inspection_date) 
	VALUES 
		-- Jan 2024
		('全聯福利中心 延吉店', '臺北市大安區延吉街', 25.0421, 121.5532, '蔬果殘留農藥', '農藥殘留', 'PASS', '2024-01-15'),
		('家樂福 桂林店', '臺北市萬華區桂林路', 25.0381, 121.5052, '肉品瘦肉精', '瘦肉精', 'PASS', '2024-01-18'),
		('某知名冰店', '臺北市大安區永康街', 25.0321, 121.5292, '冰品微生物', '微生物', 'PASS', '2024-01-20'),
		-- Feb 2024
		('濱江市場 攤商A', '臺北市中山區民族東路', 25.0668, 121.5361, '蔬果殘留農藥', '農藥殘留', 'FAIL', '2024-02-10'),
		('美廉社 內湖店', '臺北市內湖區', 25.0821, 121.5702, '肉品瘦肉精', '瘦肉精', 'PASS', '2024-02-15'),
		-- Mar 2024
		('士林夜市 冰攤', '臺北市士林區', 25.0879, 121.5248, '冰品微生物', '微生物', 'PASS', '2024-03-05'),
		('傳統市場 菜販', '臺北市北投區', 25.1312, 121.5002, '蔬果殘留農藥', '農藥殘留', 'PASS', '2024-03-12'),
		-- Apr 2024
		('鼎泰豐 信義店', '臺北市大安區信義路二段194號', 25.0334, 121.5298, '肉品瘦肉精', '瘦肉精', 'PASS', '2024-04-15'),
		('上引水產', '臺北市中山區民族東路410巷2弄18號', 25.0668, 121.5361, '冰品微生物', '微生物', 'FAIL', '2024-04-20'),
		-- May 2024
		('阜杭豆漿', '臺北市中正區忠孝東路一段108號', 25.0441, 121.5248, '蔬果殘留農藥', '農藥殘留', 'PASS', '2024-05-18'),
		('阿宗麵線', '臺北市萬華區峨眉街8-1號', 25.0433, 121.5078, '肉品瘦肉精', '瘦肉精', 'FAIL', '2024-05-29'),
		-- Jun 2024 (Summer - higher ice risk)
		('西門町 冰果室', '臺北市萬華區', 25.0431, 121.5067, '冰品微生物', '微生物', 'FAIL', '2024-06-10'),
		('芒果冰專賣店', '臺北市大安區', 25.0325, 121.5301, '冰品微生物', '微生物', 'FAIL', '2024-06-15'),
		('頂好超市', '臺北市松山區', 25.0501, 121.5775, '蔬果殘留農藥', '農藥殘留', 'PASS', '2024-06-22'),
		-- Jul 2024 (Summer - higher ice risk)
		('饒河夜市 刨冰', '臺北市松山區', 25.0505, 121.5778, '冰品微生物', '微生物', 'FAIL', '2024-07-05'),
		('寧夏夜市 豆花', '臺北市大同區', 25.0558, 121.5151, '冰品微生物', '微生物', 'FAIL', '2024-07-12'),
		('肉品批發商', '臺北市南港區', 25.0594, 121.6158, '肉品瘦肉精', '瘦肉精', 'PASS', '2024-07-20'),
		-- Aug 2024 (Summer - higher ice risk)
		('公館 撞奶', '臺北市中正區', 25.0155, 121.5333, '冰品微生物', '微生物', 'FAIL', '2024-08-10'),
		('師大夜市 飲料店', '臺北市大安區', 25.0234, 121.5298, '冰品微生物', '微生物', 'FAIL', '2024-08-18'),
		('有機蔬菜店', '臺北市文山區', 24.9877, 121.5766, '蔬果殘留農藥', '農藥殘留', 'PASS', '2024-08-25'),
		-- Sep 2024
		('百貨公司 超市', '臺北市信義區', 25.0355, 121.5655, '蔬果殘留農藥', '農藥殘留', 'PASS', '2024-09-10'),
		('知名火鍋店', '臺北市中山區', 25.0545, 121.5255, '肉品瘦肉精', '瘦肉精', 'PASS', '2024-09-22'),
		-- Oct 2024
		('大潤發', '臺北市內湖區', 25.0688, 121.5741, '肉品瘦肉精', '瘦肉精', 'PASS', '2024-10-15'),
		('連鎖簡餐店', '臺北市松山區', 25.0608, 121.5601, '蔬果殘留農藥', '農藥殘留', 'FAIL', '2024-10-25'),
		-- Nov 2024
		('南門市場', '臺北市中正區', 25.0328, 121.5181, '肉品瘦肉精', '瘦肉精', 'PASS', '2024-11-12'),
		('迪化街 商號', '臺北市大同區', 25.0566, 121.5101, '蔬果殘留農藥', '農藥殘留', 'PASS', '2024-11-20'),
		-- Dec 2024
		('信義區 饗食天堂', '臺北市信義區松壽路12號', 25.0354, 121.5654, '肉品瘦肉精', '瘦肉精', 'PASS', '2024-12-15'),
		('北投市場', '臺北市北投區', 25.1312, 121.5002, '蔬果殘留農藥', '農藥殘留', 'PASS', '2024-12-22');`

	if err := models.DBDashboard.Exec(insertDataSQL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data: " + err.Error()})
		return
	}

    // 3. Register components
    var mapCompID, trendCompID, riskCompID int64
    
    // 3a. Map Component
    err := models.DBManager.Raw("INSERT INTO components (index, name) VALUES ('food_safety', '臺北食安稽查地圖') ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name RETURNING id").Scan(&mapCompID).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register map component: " + err.Error()})
        return
    }

    // 3b. Trend Component (TimelineSeparateChart)
    err = models.DBManager.Raw("INSERT INTO components (index, name) VALUES ('food_safety_trend', '抽驗合格率趨勢') ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name RETURNING id").Scan(&trendCompID).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register trend component: " + err.Error()})
        return
    }

    // 3c. Risk Component (RadarChart)
    err = models.DBManager.Raw("INSERT INTO components (index, name) VALUES ('food_safety_risk', '食品類別風險權重') ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name RETURNING id").Scan(&riskCompID).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register risk component: " + err.Error()})
        return
    }

    // 4. Register chart configs
    // Map Config
    chartSQL := `INSERT INTO component_charts (index, color, types, unit) 
                 VALUES ('food_safety', ARRAY['#C8E6C9','#4CAF50','#1B5E20'], ARRAY['DistrictChart', 'ColumnChart', 'BarChart'], '%')
                 ON CONFLICT (index) DO UPDATE SET color = EXCLUDED.color, types = EXCLUDED.types, unit = EXCLUDED.unit`
    models.DBManager.Exec(chartSQL)

    // Trend Config
    chartSQL = `INSERT INTO component_charts (index, color, types, unit) 
                 VALUES ('food_safety_trend', ARRAY['#4CAF50'], ARRAY['TimelineSeparateChart'], '%')
                 ON CONFLICT (index) DO UPDATE SET color = EXCLUDED.color, types = EXCLUDED.types, unit = EXCLUDED.unit`
    models.DBManager.Exec(chartSQL)

    // Risk Config
    chartSQL = `INSERT INTO component_charts (index, color, types, unit) 
                 VALUES ('food_safety_risk', ARRAY['#FF5252'], ARRAY['RadarChart'], '')
                 ON CONFLICT (index) DO UPDATE SET color = EXCLUDED.color, types = EXCLUDED.types, unit = EXCLUDED.unit`
    models.DBManager.Exec(chartSQL)

    // 5. Map Config for the map component
    var mapID int
    models.DBManager.Exec("DELETE FROM component_maps WHERE index = 'food_safety'")
    mapSQL := `INSERT INTO component_maps (index, title, type, source, paint) 
               VALUES ('food_safety', '食安稽查點', 'circle', 'api', '{"circle-radius": 10, "circle-stroke-width": 2, "circle-stroke-color": "#ffffff", "circle-color": ["match", ["get", "status"], "PASS", "#4CAF50", "FAIL", "#F44336", "PENDING", "#FFC107", "#9E9E9E"]}')
               RETURNING id`
    models.DBManager.Raw(mapSQL).Scan(&mapID)

    // 6. Register Queries in query_charts
    
    // Map Query (GeoJSON) - use CTE to allow safe subquery wrapping by GetGeoJSONData
    mapQuerySQL := `WITH inspection_features AS (
                        SELECT jsonb_build_object(
                            'type', 'Feature',
                            'id', id,
                            'geometry', ST_AsGeoJSON(ST_MakePoint(lng, lat))::jsonb,
                            'properties', jsonb_build_object(
                                'name', name,
                                'address', address,
                                'status', status,
                                'category', category,
                                'test_item', test_item,
                                'inspection_date', inspection_date
                            )
                        ) AS feature
                        FROM food_safety_inspections
                    )
                    SELECT jsonb_build_object('type', 'FeatureCollection', 'features', jsonb_agg(feature))::text AS result
                    FROM inspection_features`
    
    models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'food_safety' AND city = 'taipei'")
    qcSQL := fmt.Sprintf(`INSERT INTO query_charts (index, city, query_type, query_chart, map_config_ids, source, short_desc, created_at, updated_at, time_from, time_to) 
              VALUES ('food_safety', 'taipei', 'geojson', ?, ARRAY[%d], '臺北市政府衛生局', '顯示臺北市餐廳食安稽查結果與合格狀態。', NOW(), NOW(), 'max', 'now')`, mapID)
    models.DBManager.Exec(qcSQL, mapQuerySQL)

    // Trend Query (Time series)
    trendQuerySQL := `SELECT 
                        date_trunc('month', inspection_date) as x_axis,
                        '合格率' as y_axis,
                        round(sum(case when status = 'PASS' then 1 else 0 end)::numeric / count(*)::numeric * 100, 1) as data
                    FROM food_safety_inspections
                    GROUP BY 1, 2
                    ORDER BY 1`
    
    models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'food_safety_trend' AND city = 'taipei'")
    qcSQL = `INSERT INTO query_charts (index, city, query_type, query_chart, source, short_desc, created_at, updated_at, time_from, time_to) 
              VALUES ('food_safety_trend', 'taipei', 'time', ?, '臺北市政府衛生局', '顯示每月食品抽驗合格率走勢，觀察季節性風險。', NOW(), NOW(), 'max', 'now')`
    models.DBManager.Exec(qcSQL, trendQuerySQL)

    // Risk Query (Radar/ThreeD)
    riskQuerySQL := `SELECT 
                        category as x_axis,
                        '風險權重' as y_axis,
                        round(sum(case when status = 'FAIL' then 1 else 0 end)::numeric / count(*)::numeric * 100, 1)::int as data
                    FROM food_safety_inspections
                    GROUP BY 1, 2`
    
    models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'food_safety_risk' AND city = 'taipei'")
    qcSQL = `INSERT INTO query_charts (index, city, query_type, query_chart, source, short_desc, created_at, updated_at, time_from, time_to) 
              VALUES ('food_safety_risk', 'taipei', 'three_d', ?, '臺北市政府衛生局', '顯示不同食品類別的風險權重（不合格率）。', NOW(), NOW(), 'max', 'now')`
    models.DBManager.Exec(qcSQL, riskQuerySQL)

    // 7. Create/Update Dashboard
    var dashID int
    dashSQL := `INSERT INTO dashboards (index, name, components, icon, created_at, updated_at) 
                VALUES ('food_safety', '食安地圖與趨勢', ARRAY[?, ?, ?]::integer[], 'restaurant', NOW(), NOW())
                ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name, components = EXCLUDED.components, icon = EXCLUDED.icon, updated_at = NOW()
                RETURNING id`
    err = models.DBManager.Raw(dashSQL, mapCompID, trendCompID, riskCompID).Scan(&dashID).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dashboard: " + err.Error()})
        return
    }

    // 8. Assign to taipei group (group_id=2)
    models.DBManager.Exec("DELETE FROM dashboard_groups WHERE dashboard_id = ?", dashID)
    models.DBManager.Exec("INSERT INTO dashboard_groups (dashboard_id, group_id) VALUES (?, 2)", dashID)

	c.JSON(http.StatusOK, gin.H{
        "message": "Food safety initialization successful with trends and alerts", 
        "components": []int64{mapCompID, trendCompID, riskCompID}, 
        "dashboard_id": dashID,
    })
}
