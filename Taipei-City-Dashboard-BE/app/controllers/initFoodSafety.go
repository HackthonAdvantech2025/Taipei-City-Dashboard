package controllers

import (
	"TaipeiCityDashboardBE/app/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitFoodSafety(c *gin.Context) {
	// 1. Create table and insert data in dashboard DB
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

	// 2. Fetch real data from Taipei Data Platform
	// Skipped: We are now using real data imported via CSV, so we do not TRUNCATE or generate fake data here.

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
    mapSQL := `INSERT INTO component_maps (index, title, type, source, paint) 
               VALUES ('food_safety', '食安稽查點', 'circle', 'api', '{"circle-radius": 10, "circle-stroke-width": 2, "circle-stroke-color": "#ffffff", "circle-color": ["match", ["get", "status"], "PASS", "#4CAF50", "FAIL", "#F44336", "PENDING", "#FFC107", "#9E9E9E"]}')
               ON CONFLICT (index) DO UPDATE SET paint = EXCLUDED.paint
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
                    SELECT jsonb_build_object('type', 'FeatureCollection', 'features', COALESCE(jsonb_agg(feature), '[]'::jsonb))::text AS result
                    FROM inspection_features`
    
    models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'food_safety' AND city = 'taipei'")
    qcSQL := fmt.Sprintf(`INSERT INTO query_charts (index, city, query_type, query_chart, map_config_ids, source, short_desc, created_at, updated_at, time_from, time_to) 
              VALUES ('food_safety', 'taipei', 'geojson', ?, ARRAY[%d], '臺北市政府衛生局', '顯示臺北市餐廳食安稽查結果與合格狀態。', NOW(), NOW(), '2026-01-01', '2026-04-30')`, mapID)
    models.DBManager.Exec(qcSQL, mapQuerySQL)

    // Trend Query (Time series)
    trendQuerySQL := `SELECT 
                        date_trunc('month', inspection_date) as x_axis,
                        '合格率' as y_axis,
                        round(sum(case when status = 'PASS' then 1 else 0 end)::numeric / nullif(count(*), 0)::numeric * 100, 1) as data
                    FROM food_safety_inspections
                    GROUP BY 1, 2
                    ORDER BY 1`
    
    models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'food_safety_trend' AND city = 'taipei'")
    qcSQL = `INSERT INTO query_charts (index, city, query_type, query_chart, source, short_desc, created_at, updated_at, time_from, time_to) 
              VALUES ('food_safety_trend', 'taipei', 'time', ?, '臺北市政府衛生局', '顯示每月食品抽驗合格率走勢，觀察季節性風險。', NOW(), NOW(), '2026-01-01', '2026-04-30')`
    models.DBManager.Exec(qcSQL, trendQuerySQL)

    // Risk Query (Radar/ThreeD) - Composite Risk Scoring Model
    riskQuerySQL := `WITH CategorizedData AS (
                        SELECT 
                            CASE 
                                WHEN category LIKE '%水產%' OR category LIKE '%生食%' OR category LIKE '%蟹%' OR category LIKE '%魚%' THEN '生食海鮮 (高風險)'
                                WHEN category LIKE '%肉%' OR category LIKE '%蛋%' OR category LIKE '%奶油%' THEN '肉類乳製 (高風險)'
                                WHEN category LIKE '%菜%' OR category LIKE '%果%' OR category LIKE '%植物%' OR category LIKE '%蘑菇%' OR category LIKE '%木耳%' THEN '生鮮蔬果 (中高風險)'
                                WHEN category LIKE '%飲%' OR category LIKE '%冰%' OR category LIKE '%豆%' THEN '飲冰品豆類 (中風險)'
                                WHEN category LIKE '%麵%' OR category LIKE '%粉%' OR category LIKE '%乾%' OR category LIKE '%包裝%' OR category LIKE '%蜜餞%' OR category LIKE '%月餅%' THEN '乾貨烘焙 (低風險)'
                                ELSE '其他加工品'
                            END as main_category,
                            status
                        FROM food_safety_inspections
                        WHERE category NOT IN ('定期稽查', '最新稽查', 'HACCP稽查')
                    )
                    SELECT 
                        main_category as x_axis,
                        '風險綜合評分' as y_axis,
                        round(
                            (
                                -- 實際違規率佔 40%
                                (sum(case when status = 'FAIL' then 1 else 0 end)::numeric / nullif(count(*), 0)::numeric * 40)
                                + 
                                -- 基礎嚴重度權重佔 60%
                                (
                                    CASE 
                                        WHEN main_category = '生食海鮮 (高風險)' THEN 95
                                        WHEN main_category = '肉類乳製 (高風險)' THEN 85
                                        WHEN main_category = '生鮮蔬果 (中高風險)' THEN 75
                                        WHEN main_category = '飲冰品豆類 (中風險)' THEN 50
                                        WHEN main_category = '乾貨烘焙 (低風險)' THEN 25
                                        ELSE 40
                                    END * 0.6
                                )
                            ), 0
                        )::int as data
                    FROM CategorizedData
                    GROUP BY main_category`
    
    models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'food_safety_risk' AND city = 'taipei'")
    qcSQL = `INSERT INTO query_charts (index, city, query_type, query_chart, source, short_desc, created_at, updated_at, time_from, time_to) 
              VALUES ('food_safety_risk', 'taipei', 'three_d', ?, '臺北市政府衛生局', '顯示不同食品類別的風險權重（不合格率）。', NOW(), NOW(), '2026-01-01', '2026-04-30')`
    models.DBManager.Exec(qcSQL, riskQuerySQL)

    // 7. Create/Update Dashboard
    var dashID int
    dashSQL := `INSERT INTO dashboards (index, name, components, icon, created_at, updated_at) 
                VALUES ('food_safety', '食安地圖與趨勢', ARRAY[?, ?, ?]::integer[], 'restaurant', NOW(), NOW())
                ON CONFLICT (index) DO UPDATE SET 
                    name = EXCLUDED.name, 
                    components = ARRAY(SELECT DISTINCT unnest(dashboards.components || EXCLUDED.components)), 
                    icon = EXCLUDED.icon, 
                    updated_at = NOW()
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
