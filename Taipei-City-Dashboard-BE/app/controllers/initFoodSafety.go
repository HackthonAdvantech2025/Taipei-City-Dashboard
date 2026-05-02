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
		status VARCHAR(50),
		inspection_date DATE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`
	
	if err := models.DBDashboard.Exec(createTableSQL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create table: " + err.Error()})
		return
	}
	// Clear existing data to ensure exactly 50 rows
	models.DBDashboard.Exec("TRUNCATE TABLE food_safety_inspections")

	insertDataSQL := `
	INSERT INTO food_safety_inspections (name, address, lat, lng, status, inspection_date) 
	SELECT * FROM (VALUES 
		('鼎泰豐 信義店', '臺北市大安區信義路二段194號', 25.0334, 121.5298, 'PASS', '2024-04-15'::DATE),
		('馬辣', '臺北市萬華區西寧南路157號', 25.0431, 121.5067, 'PASS', '2024-04-10'::DATE),
		('上引水產', '臺北市中山區民族東路410巷2弄18號', 25.0668, 121.5361, 'FAIL', '2024-04-20'::DATE),
		('饒河夜市', '臺北市松山區饒河街', 25.0501, 121.5775, 'PENDING', '2024-04-22'::DATE),
		('阜杭豆漿', '臺北市中正區忠孝東路一段108號', 25.0441, 121.5248, 'PASS', '2024-04-18'::DATE),
		('士林夜市美食區', '臺北市士林區基河路101號', 25.0879, 121.5248, 'FAIL', '2024-04-25'::DATE),
		('金泰日式料理', '臺北市內湖區舊宗路二段121巷34號', 25.0688, 121.5741, 'PASS', '2024-04-26'::DATE),
		('欣葉小聚 南港', '臺北市南港區經貿二路166號', 25.0594, 121.6158, 'FAIL', '2024-04-27'::DATE),
		('微熱山丘 台北', '臺北市松山區民生東路五段36巷4弄1號', 25.0608, 121.5601, 'PASS', '2024-04-28'::DATE),
		('阿宗麵線 西門', '臺北市萬華區峨眉街8-1號', 25.0433, 121.5078, 'FAIL', '2024-04-29'::DATE),
		('北投滿來拉麵', '臺北市北投區中山路1-8號', 25.1374, 121.5055, 'PASS', '2024-04-15'::DATE),
		('北投市場大興肉羹', '臺北市北投區新市街30號', 25.1312, 121.5002, 'PASS', '2024-04-16'::DATE),
		('士林阿輝麵線', '臺北市士林區大南路84號', 25.0888, 121.5255, 'FAIL', '2024-04-17'::DATE),
		('內湖小蒙牛', '臺北市內湖區內湖路一段248號', 25.0823, 121.5711, 'PASS', '2024-04-18'::DATE),
		('內湖覺旅咖啡', '臺北市內湖區瑞光路583巷24號', 25.0821, 121.5702, 'PASS', '2024-04-19'::DATE),
		('中山區雞家莊', '臺北市中山區長春路55號', 25.0545, 121.5255, 'FAIL', '2024-04-20'::DATE),
		('中山肥前屋', '臺北市中山區中山北路一段121巷13號', 25.0498, 121.5222, 'PASS', '2024-04-21'::DATE),
		('大同區民樂旗魚吐', '臺北市大同區民樂街3號', 25.0541, 121.5105, 'PASS', '2024-04-22'::DATE),
		('大同區意麵王', '臺北市大同區歸綏街204號', 25.0574, 121.5122, 'FAIL', '2024-04-23'::DATE),
		('中正區金峰魯肉飯', '臺北市中正區羅斯福路一段10號', 25.0321, 121.5188, 'PASS', '2024-04-24'::DATE),
		('中正區張家食堂', '臺北市中正區忠孝西路一段47號', 25.0478, 121.5171, 'PASS', '2024-04-25'::DATE),
		('萬華兩喜號魷魚焿', '臺北市萬華區廣州街245號', 25.0366, 121.4998, 'PASS', '2024-04-26'::DATE),
		('萬華龍都冰果專業家', '臺北市萬華區廣州街168號', 25.0364, 121.5002, 'PASS', '2024-04-27'::DATE),
		('大安區京兆尹', '臺北市大安區仁愛路四段33號', 25.0388, 121.5455, 'PASS', '2024-04-28'::DATE),
		('大安區銀翼餐廳', '臺北市大安區金山南路二段18號', 25.0322, 121.5277, 'PASS', '2024-04-29'::DATE),
		('信義區糖朝', '臺北市信義區忠孝東路四段197號', 25.0415, 121.5511, 'PASS', '2024-04-30'::DATE),
		('信義區牛角燒肉', '臺北市信義區松壽路12號', 25.0355, 121.5655, 'FAIL', '2024-05-01'::DATE),
		('松山區富錦樹台菜', '臺北市松山區敦化北路199巷17號', 25.0588, 121.5502, 'PASS', '2024-05-02'::DATE),
		('南港區漉海鮮蒸氣鍋', '臺北市南港區經貿二路188號', 25.0581, 121.6166, 'PASS', '2024-05-03'::DATE),
		('文山區阿義師大茶壺', '臺北市文山區指南路三段38巷37-1號', 24.9691, 121.5891, 'PASS', '2024-05-04'::DATE),
		('文山區龍門客棧', '臺北市文山區指南路三段38巷22-2號', 24.9681, 121.5877, 'PASS', '2024-05-05'::DATE),
		('北投蓬萊餐廳', '臺北市北投區中和街238號', 25.1377, 121.5011, 'PASS', '2024-05-06'::DATE),
		('士林富樂台式刷刷鍋', '臺北市士林區承德路四段75-1號', 25.0844, 121.5241, 'PASS', '2024-05-07'::DATE),
		('中山區點水樓', '臺北市中山區樂群三路299號', 25.0833, 121.5588, 'PASS', '2024-05-08'::DATE),
		('大同區寧夏夜市', '臺北市大同區寧夏路', 25.0558, 121.5151, 'PASS', '2024-05-09'::DATE),
		('中正區公館陳三鼎', '臺北市中正區羅斯福路三段316巷8弄2號', 25.0155, 121.5333, 'PASS', '2024-05-10'::DATE),
		('萬華區小王煮瓜', '臺北市萬華區華西街17之4號', 25.0371, 121.4988, 'PASS', '2024-05-11'::DATE),
		('大安區度小月', '臺北市大安區忠孝東路四段216巷8弄12號', 25.0411, 121.5522, 'PASS', '2024-05-12'::DATE),
		('信義區饗食天堂', '臺北市信義區松壽路12號', 25.0354, 121.5654, 'FAIL', '2024-05-13'::DATE),
		('松山區小巨蛋美食', '臺北市松山區南京東路四段2號', 25.0511, 121.5501, 'PASS', '2024-05-14'::DATE),
		('內湖區泰里亞', '臺北市內湖區瑞光路513巷30號', 25.0799, 121.5722, 'PASS', '2024-05-15'::DATE),
		('南港區寒舍樂廚', '臺北市南港區經貿二路1號', 25.0566, 121.6144, 'PASS', '2024-05-16'::DATE),
		('文山區政大校園美食', '臺北市文山區指南路二段64號', 24.9877, 121.5766, 'PASS', '2024-05-17'::DATE),
		('大同區迪化街大稻埕', '臺北市大同區迪化街一段', 25.0566, 121.5101, 'PASS', '2024-05-18'::DATE),
		('中正區南門市場', '臺北市中正區羅斯福路一段8號', 25.0328, 121.5181, 'PASS', '2024-05-19'::DATE),
		('萬華區華西街夜市', '臺北市萬華區華西街', 25.0381, 121.4991, 'PASS', '2024-05-20'::DATE),
		('士林區天母美食', '臺北市士林區中山北路六段', 25.1122, 121.5288, 'PASS', '2024-05-21'::DATE),
		('北投區石牌美食', '臺北市北投區裕民六路', 25.1155, 121.5144, 'PASS', '2024-05-22'::DATE),
		('內湖區西湖市場', '臺北市內湖區內湖路一段285號', 25.0822, 121.5677, 'PASS', '2024-05-23'::DATE)
	) AS t (name, address, lat, lng, status, inspection_date)
	WHERE NOT EXISTS (SELECT 1 FROM food_safety_inspections WHERE name = t.name);`

	if err := models.DBDashboard.Exec(insertDataSQL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data: " + err.Error()})
		return
	}

    // 3. Register component and get ID
    var compID int64
    err := models.DBManager.Raw("INSERT INTO components (index, name) VALUES ('food_safety', '臺北食安稽查地圖') ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name RETURNING id").Scan(&compID).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register component: " + err.Error()})
        return
    }

    // 3b. Register component_charts (REQUIRED - createTempComponentDB does INNER JOIN on this table)
    // Using a pure Green gradient: Light Green -> Dark Green
    chartSQL := `INSERT INTO component_charts (index, color, types, unit) 
                 VALUES ('food_safety', ARRAY['#C8E6C9','#4CAF50','#1B5E20'], ARRAY['DistrictChart', 'ColumnChart'], '%')
                 ON CONFLICT (index) DO UPDATE SET color = EXCLUDED.color, types = EXCLUDED.types, unit = EXCLUDED.unit`
    if err = models.DBManager.Exec(chartSQL).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register component_charts: " + err.Error()})
        return
    }

    // 4. Register map config
    var mapID int
    models.DBManager.Exec("DELETE FROM component_maps WHERE index = 'food_safety'")
    mapSQL := `INSERT INTO component_maps (index, title, type, source, paint) 
               VALUES ('food_safety', '食安稽查點', 'circle', 'api', '{"circle-radius": 10, "circle-stroke-width": 2, "circle-stroke-color": "#ffffff", "circle-color": ["match", ["get", "status"], "PASS", "#4CAF50", "FAIL", "#F44336", "PENDING", "#FFC107", "#9E9E9E"]}')
               RETURNING id`
    err = models.DBManager.Raw(mapSQL).Scan(&mapID).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register map_config: " + err.Error()})
        return
    }

    // d. Query Charts
    querySQL := `SELECT jsonb_build_object('type', 'FeatureCollection', 'features', jsonb_agg(features.feature))
                 FROM (SELECT jsonb_build_object('type', 'Feature', 'id', id, 'geometry', ST_AsGeoJSON(ST_MakePoint(lng, lat))::jsonb,
                 'properties', jsonb_build_object('name', name, 'address', address, 'status', status, 'inspection_date', inspection_date)) AS feature
                 FROM food_safety_inspections) features`
    
    models.DBManager.Exec("DELETE FROM query_charts WHERE index = 'food_safety' AND city = 'taipei'")
    qcSQL := fmt.Sprintf(`INSERT INTO query_charts (index, city, query_type, query_chart, map_config_ids, source, short_desc, created_at, updated_at) 
              VALUES ('food_safety', 'taipei', 'geojson', ?, ARRAY[%d], '臺北市政府衛生局', '顯示臺北市餐廳食安稽查結果與合格狀態。', NOW(), NOW())`, mapID)
    
    if err := models.DBManager.Exec(qcSQL, querySQL).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register query_chart: " + err.Error()})
        return
    }

    // e. Create Dashboard
    var dashID int
    dashSQL := `INSERT INTO dashboards (index, name, components, icon, created_at, updated_at) 
                VALUES ('food_safety', '食安地圖', ARRAY[?]::integer[], 'restaurant', NOW(), NOW())
                ON CONFLICT (index) DO UPDATE SET name = EXCLUDED.name, components = EXCLUDED.components, icon = EXCLUDED.icon, updated_at = NOW()
                RETURNING id`
    err = models.DBManager.Raw(dashSQL, compID).Scan(&dashID).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dashboard: " + err.Error()})
        return
    }

    // f. Assign to taipei group (group_id=2) so it shows in the city sidebar
    models.DBManager.Exec("DELETE FROM dashboard_groups WHERE dashboard_id = ?", dashID)
    err = models.DBManager.Exec("INSERT INTO dashboard_groups (dashboard_id, group_id) VALUES (?, 2)", dashID).Error
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign dashboard group: " + err.Error()})
        return
    }

	c.JSON(http.StatusOK, gin.H{"message": "Food safety initialization successful", "component_id": compID, "dashboard_id": dashID})
}
