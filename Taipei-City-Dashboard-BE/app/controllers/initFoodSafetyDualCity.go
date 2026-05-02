package controllers

import (
	"TaipeiCityDashboardBE/app/models"
	"net/http"
	"github.com/gin-gonic/gin"
)

func InitFoodSafetyDualCity(c *gin.Context) {
	// 1. Create table fact_food_safety_cases
	models.DBDashboard.Exec("DROP TABLE IF EXISTS fact_food_safety_cases")
	
	createTableSQL := `
	CREATE TABLE fact_food_safety_cases (
		id SERIAL PRIMARY KEY,
		city VARCHAR(50),
		case_date DATE,
		case_type VARCHAR(100),
		business_type VARCHAR(100),
		violation_level VARCHAR(50),
		penalty_amount INT,
		source VARCHAR(50)
	);`
	
	if err := models.DBDashboard.Exec(createTableSQL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create table: " + err.Error()})
		return
	}

	// 2. Insert mock data for Taipei and New Taipei
	insertDataSQL := `
	INSERT INTO fact_food_safety_cases (city, case_date, case_type, business_type, violation_level, penalty_amount, source) 
	VALUES 
		-- 台北市 (Taipei)
		('臺北市', '2024-03-10', '過期食品', '超市大賣場', '輕微', 30000, '稽查'),
		('臺北市', '2024-03-15', '環境衛生不佳', '夜市攤販', '中等', 60000, '檢舉'),
		('臺北市', '2024-04-05', '標示不實', '連鎖早餐店', '輕微', 30000, '抽驗'),
		('臺北市', '2024-04-12', '食物中毒', '高級餐廳', '嚴重', 200000, '檢舉'),
		('臺北市', '2024-04-20', '農藥殘留', '手搖飲料', '中等', 60000, '抽驗'),
		('臺北市', '2024-05-01', '環境衛生不佳', '連鎖早餐店', '嚴重', 120000, '稽查'),
		('臺北市', '2024-05-15', '非法添加物', '夜市攤販', '嚴重', 150000, '抽驗'),
		
		-- 新北市 (New Taipei)
		('新北市', '2024-03-12', '標示不實', '超級市場', '輕微', 30000, '稽查'),
		('新北市', '2024-03-18', '農藥殘留', '傳統市場', '嚴重', 100000, '抽驗'),
		('新北市', '2024-04-02', '環境衛生不佳', '一般餐廳', '中等', 60000, '檢舉'),
		('新北市', '2024-04-10', '過期食品', '夜市攤販', '中等', 60000, '稽查'),
		('新北市', '2024-04-25', '非法添加物', '食品工廠', '嚴重', 300000, '抽驗'),
		('新北市', '2024-05-05', '保存溫度異常', '超級市場', '輕微', 30000, '稽查'),
		('新北市', '2024-05-18', '食物中毒', '學校團膳', '嚴重', 500000, '檢舉');
	`
    // Multiply data for better chart visualization
    models.DBDashboard.Exec(insertDataSQL)
	for i := 0; i < 5; i++ {
		models.DBDashboard.Exec(`
			INSERT INTO fact_food_safety_cases (city, case_date, case_type, business_type, violation_level, penalty_amount, source)
			SELECT city, case_date + (random() * 60)::int, case_type, business_type, violation_level, penalty_amount, source
			FROM fact_food_safety_cases LIMIT 14;
		`)
	}

    // 3. Register components
    var cityCompID, bizCompID, penaltyCompID int64
    
    // 3a. Dual City Comparison (ColumnChart)
    models.DBManager.Raw("INSERT INTO components (index, name) VALUES ('dual_city_cases', '雙北案件數比較') ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name RETURNING id").Scan(&cityCompID)
    
    // 3b. Business Type (DonutChart)
    models.DBManager.Raw("INSERT INTO components (index, name) VALUES ('dual_city_biz_type', '違規業別分佈') ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name RETURNING id").Scan(&bizCompID)

    // 3c. Penalty Amount Trend (TimelineSeparateChart)
    models.DBManager.Raw("INSERT INTO components (index, name) VALUES ('dual_city_penalty_trend', '每月裁罰總金額趨勢') ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name RETURNING id").Scan(&penaltyCompID)

    // 4. Register chart configs
    models.DBManager.Exec(`INSERT INTO component_charts (index, color, types, unit) 
                 VALUES ('dual_city_cases', ARRAY['#2196F3', '#FF9800'], ARRAY['ColumnChart'], '件')
                 ON CONFLICT (index) DO UPDATE SET color = EXCLUDED.color, types = EXCLUDED.types, unit = EXCLUDED.unit`)

    models.DBManager.Exec(`INSERT INTO component_charts (index, color, types, unit) 
                 VALUES ('dual_city_biz_type', ARRAY['#4CAF50','#E91E63','#9C27B0','#00BCD4','#795548','#607D8B'], ARRAY['DonutChart'], '件')
                 ON CONFLICT (index) DO UPDATE SET color = EXCLUDED.color, types = EXCLUDED.types, unit = EXCLUDED.unit`)

    models.DBManager.Exec(`INSERT INTO component_charts (index, color, types, unit) 
                 VALUES ('dual_city_penalty_trend', ARRAY['#F44336', '#3F51B5'], ARRAY['TimelineSeparateChart'], '元')
                 ON CONFLICT (index) DO UPDATE SET color = EXCLUDED.color, types = EXCLUDED.types, unit = EXCLUDED.unit`)

    // 5. Register Queries in query_charts
    
    // City Comparison Query (two_d)
    cityQuery := `SELECT 
                        case_type as x_axis,
                        city as y_axis,
                        count(*) as data
                    FROM fact_food_safety_cases
                    GROUP BY 1, 2
					ORDER BY 3 DESC`
    
    models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'dual_city_cases' AND city = 'taipei'")
    models.DBManager.Exec(`INSERT INTO query_charts (index, city, query_type, query_chart, source, short_desc, created_at, updated_at, time_from, time_to) 
              VALUES ('dual_city_cases', 'taipei', 'two_d', ?, '雙北市政府開放資料', '比較台北市與新北市在不同食安違規案件類型的數量。', NOW(), NOW(), 'max', 'now')`, cityQuery)

    // Business Type Query (two_d)
    bizQuery := `SELECT 
                        business_type as x_axis,
                        count(*) as data
                    FROM fact_food_safety_cases
                    GROUP BY 1
					ORDER BY 2 DESC`
    
    models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'dual_city_biz_type' AND city = 'taipei'")
    models.DBManager.Exec(`INSERT INTO query_charts (index, city, query_type, query_chart, source, short_desc, created_at, updated_at, time_from, time_to) 
              VALUES ('dual_city_biz_type', 'taipei', 'two_d', ?, '雙北市政府開放資料', '統計最常發生食安問題的營業場所類型。', NOW(), NOW(), 'max', 'now')`, bizQuery)

	// Penalty Trend Query (time)
	penaltyQuery := `SELECT 
						date_trunc('month', case_date) as x_axis,
						city as y_axis,
						sum(penalty_amount) as data
					FROM fact_food_safety_cases
					GROUP BY 1, 2
					ORDER BY 1`
	
	models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'dual_city_penalty_trend' AND city = 'taipei'")
	models.DBManager.Exec(`INSERT INTO query_charts (index, city, query_type, query_chart, source, short_desc, created_at, updated_at, time_from, time_to) 
			  VALUES ('dual_city_penalty_trend', 'taipei', 'time', ?, '雙北市政府開放資料', '追蹤雙北每月開出的食安裁罰總金額走勢。', NOW(), NOW(), '1y', 'now')`, penaltyQuery)

    // 6. Create Dashboard
    var dashID int
    dashSQL := `INSERT INTO dashboards (index, name, components, icon, created_at, updated_at) 
                VALUES ('food_safety_dual_city', '雙北食安跨區域分析', ARRAY[?, ?, ?]::integer[], 'compare_arrows', NOW(), NOW())
                ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name, components = EXCLUDED.components, icon = EXCLUDED.icon, updated_at = NOW()
                RETURNING id`
    models.DBManager.Raw(dashSQL, cityCompID, bizCompID, penaltyCompID).Scan(&dashID)

    // 7. Assign to taipei group (group_id=2)
    models.DBManager.Exec("DELETE FROM dashboard_groups WHERE dashboard_id = ?", dashID)
    models.DBManager.Exec("INSERT INTO dashboard_groups (dashboard_id, group_id) VALUES (?, 2)", dashID)

	c.JSON(http.StatusOK, gin.H{
        "message": "Dual city food safety dashboard initialized successfully", 
        "dashboard_id": dashID,
    })
}
