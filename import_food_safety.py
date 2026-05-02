import urllib.request
import json
import random

url = "https://data.taipei/api/v1/dataset/3fc106cb-c8aa-4f74-8dbc-3272c7ffcae0?scope=resourceAquire&limit=100"
req = urllib.request.Request(url)
with urllib.request.urlopen(req) as response:
    data = json.loads(response.read().decode())

districts = {
    "中山區": (25.0685, 121.5282),
    "大安區": (25.0263, 121.5434),
    "松山區": (25.0565, 121.5674),
    "信義區": (25.0324, 121.5675),
    "中正區": (25.0322, 121.5173),
    "萬華區": (25.0335, 121.4988),
    "大同區": (25.0628, 121.5113),
    "內湖區": (25.0830, 121.5912),
    "南港區": (25.0551, 121.6171),
    "士林區": (25.0922, 121.5245),
    "北投區": (25.1321, 121.4987),
    "文山區": (24.9888, 121.5752)
}

sql_values = []
for item in data['result']['results']:
    # Parse name and address from "旺旺水果行(龍江水果行)/臺北市中山區龍江路189號"
    loc = item.get("抽驗地點", "")
    parts = loc.split("/")
    name = parts[0].replace("'", "''")
    address = parts[1].replace("'", "''") if len(parts) > 1 else loc.replace("'", "''")
    
    # Extract district
    lat, lng = 25.0330, 121.5654 # default
    for d, coords in districts.items():
        if d in address:
            # Add small random jitter so points don't overlap completely
            lat = coords[0] + random.uniform(-0.01, 0.01)
            lng = coords[1] + random.uniform(-0.01, 0.01)
            break
            
    category = item.get("分類", "").replace("'", "''")
    test_item = item.get("檢體名稱", "").replace("'", "''")
    status = "FAIL" # Since this dataset is "不合格清冊"
    
    # Parse date "20240102" -> "2024-01-02"
    d_raw = item.get("抽驗日期", "20240101")
    if len(d_raw) == 8:
        date_str = f"{d_raw[0:4]}-{d_raw[4:6]}-{d_raw[6:8]}"
    else:
        date_str = "2024-01-01"
        
    sql_values.append(f"('{name}', '{address}', {lat}, {lng}, '{category}', '{test_item}', '{status}', '{date_str}')")

if sql_values:
    # First clear existing mock data that is FAIL, or just insert new ones
    print(f"INSERT INTO food_safety_inspections (name, address, lat, lng, category, test_item, status, inspection_date) VALUES")
    print(",\n".join(sql_values) + ";")
