"""
iMAP 食安爬蟲 v4 — 正式版
URL 格式：
  https://imap.health.gov.tw/App_Prog/ListFood.aspx?ftype=B&area=0101&ds=20260101&de=20260630

參數：
  ftype=B   食品類
  area      行政區代碼
  ds        開始日期 yyyymmdd
  de        結束日期 yyyymmdd
  page      分頁（從1開始）
"""

import requests
from bs4 import BeautifulSoup
import pandas as pd
import time
import re
import urllib3
urllib3.disable_warnings()

BASE_URL  = "https://imap.health.gov.tw"
LIST_URL  = f"{BASE_URL}/App_Prog/ListFood.aspx"

FTYPE      = "B"
DATE_START = "20260101"
DATE_END   = "20260630"

DISTRICTS = {
    "松山區": "0101",
    "信義區": "0117",
    "大安區": "0102",
    "中山區": "0110",
    "中正區": "0118",
    "大同區": "0109",
    "萬華區": "0119",
    "文山區": "0120",
    "南港區": "0112",
    "內湖區": "0111",
    "士林區": "0115",
    "北投區": "0116",
}

HEADERS = {
    "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
    "Accept-Language": "zh-TW,zh;q=0.9",
    "Referer": BASE_URL,
}
session = requests.Session()
session.headers.update(HEADERS)
session.verify = False


# ──────────────────────────────────────────────
# 解析一頁的資料
# ──────────────────────────────────────────────
def parse_page(html: str, district_name: str) -> list[dict]:
    soup = BeautifulSoup(html, "html.parser")
    rows = []

    # ── 先存第一頁 HTML 供 debug ──
    # （只有第一次呼叫時需要，之後可移除）

    # 找所有可能包含店家資訊的區塊
    # 截圖顯示是卡片式，每筆有：店名、地址、電話、登錄字號、稽查結果、日期
    
    # 策略1：找含有「登錄字號」的 div 區塊
    blocks = soup.find_all(lambda tag: tag.name in ["div","li","article"] and 
                           "登錄字號" in tag.get_text())
    
    if not blocks:
        # 策略2：找所有連結到 DetailFood.aspx 的 <a>
        links = soup.find_all("a", href=re.compile(r"DetailFood\.aspx", re.I))
        blocks = [a.find_parent(["div","li","article"]) for a in links if a.find_parent(["div","li","article"])]
        blocks = list(dict.fromkeys(blocks))  # 去重

    if not blocks:
        # 策略3：找所有表格行
        for table in soup.find_all("table"):
            ths = [th.get_text(strip=True) for th in table.find_all("th")]
            if not ths:
                continue
            for tr in table.find_all("tr")[1:]:
                tds = tr.find_all("td")
                if not tds:
                    continue
                row = {"行政區": district_name}
                for i, td in enumerate(tds):
                    col = ths[i] if i < len(ths) else f"欄{i}"
                    row[col] = td.get_text(strip=True)
                link = tr.find("a", href=re.compile(r"Detail", re.I))
                if link:
                    row["detail_url"] = BASE_URL + "/" + link["href"].lstrip("/")
                if any(v for v in row.values() if v and v != district_name):
                    rows.append(row)

    for block in blocks:
        row = {"行政區": district_name}
        text = block.get_text("\n", strip=True)

        # 店名：通常是最大的文字 or <h> 標籤
        for tag in ["h2","h3","h4","h5","strong","b"]:
            el = block.find(tag)
            if el and len(el.get_text(strip=True)) > 1:
                row["店名"] = el.get_text(strip=True)
                break
        if "店名" not in row:
            link = block.find("a")
            if link:
                row["店名"] = link.get_text(strip=True)

        # 地址
        addr_m = re.search(r"(?:聯絡地址[：:]?\s*)([\S ]+)", text)
        row["地址"] = addr_m.group(1).strip() if addr_m else ""

        # 電話
        tel_m = re.search(r"(?:公司電話|電話)[：:]?\s*([\d\-\(\) ]+)", text)
        row["電話"] = tel_m.group(1).strip() if tel_m else ""

        # 登錄字號
        reg_m = re.search(r"(?:登錄字號)[：:]?\s*([A-Z0-9\-]+)", text)
        row["登錄字號"] = reg_m.group(1).strip() if reg_m else ""

        # 稽查結果 + 日期（如「複查不合格 2025.11.28」）
        result_m = re.search(r"(複查不合格|不合格|合格|停業|勒令)[\s\S]{0,10}?(\d{4}[.\-/]\d{1,2}[.\-/]\d{1,2})", text)
        if result_m:
            row["稽查結果"] = result_m.group(1).strip()
            row["稽查日期"] = result_m.group(2).strip().replace("/","-").replace(".","-")
        else:
            # 只找日期
            date_m = re.search(r"(\d{4}[.\-/]\d{1,2}[.\-/]\d{1,2})", text)
            row["稽查日期"] = date_m.group(1).replace("/","-").replace(".","-") if date_m else ""
            # 只找結果
            res_m = re.search(r"(複查不合格|不合格|合格|停業|勒令)", text)
            row["稽查結果"] = res_m.group(1) if res_m else "不合格"

        # detail URL
        detail_link = block.find("a", href=re.compile(r"Detail", re.I))
        if detail_link:
            href = detail_link["href"]
            row["detail_url"] = BASE_URL + "/" + href.lstrip("/") if not href.startswith("http") else href

        if row.get("店名") or row.get("登錄字號"):
            rows.append(row)

    return rows


# ──────────────────────────────────────────────
# 取得總頁數
# ──────────────────────────────────────────────
def get_total_pages(html: str) -> int:
    soup = BeautifulSoup(html, "html.parser")
    text = soup.get_text()

    # 「共 N 頁」
    m = re.search(r"共\s*(\d+)\s*頁", text)
    if m:
        return int(m.group(1))

    # 分頁下拉
    for sel in soup.find_all("select"):
        opts = [o.get("value","") for o in sel.find_all("option") if o.get("value","").isdigit()]
        if opts:
            return max(int(o) for o in opts)

    # 分頁數字連結
    nums = []
    for a in soup.find_all("a", href=re.compile(r"page=\d+", re.I)):
        m2 = re.search(r"page=(\d+)", a["href"])
        if m2:
            nums.append(int(m2.group(1)))
    # 也找純數字的 <a>
    for a in soup.find_all("a"):
        t = a.get_text(strip=True)
        if t.isdigit():
            nums.append(int(t))
    if nums:
        return max(nums)

    return 1


# ──────────────────────────────────────────────
# 爬一個行政區
# ──────────────────────────────────────────────
def scrape_district(district_name: str, area_code: str) -> list[dict]:
    all_rows = []
    page = 1

    while True:
        params = {
            "ftype": FTYPE,
            "area":  area_code,
            "ds":    DATE_START,
            "de":    DATE_END,
            "page":  page,
        }
        try:
            r = session.get(LIST_URL, params=params, timeout=20)
            r.raise_for_status()
        except Exception as e:
            print(f"    ❌ 錯誤：{e}")
            break

        # 第一頁存 debug
        if page == 1:
            with open(f"debug_{area_code}_p1.html", "w", encoding="utf-8") as f:
                f.write(r.text)

        total_pages = get_total_pages(r.text)
        rows = parse_page(r.text, district_name)
        print(f"    頁 {page}/{total_pages}：{len(rows)} 筆")

        all_rows.extend(rows)

        if page >= total_pages or len(rows) == 0:
            break

        page += 1
        time.sleep(0.6)

    return all_rows


# ──────────────────────────────────────────────
# 主程式
# ──────────────────────────────────────────────
def main():
    all_data = []

    print(f"爬取期間：{DATE_START} ～ {DATE_END}")
    print(f"目標：台北市 {len(DISTRICTS)} 個行政區\n")

    for name, code in DISTRICTS.items():
        print(f"▶ {name} ({code})")
        rows = scrape_district(name, code)
        print(f"  小計 {len(rows)} 筆")
        all_data.extend(rows)
        time.sleep(1.0)

    if not all_data:
        print("\n⚠️ 0 筆資料！請上傳 debug_0101_p1.html 給 Claude")
        return

    df = pd.DataFrame(all_data)

    # 標準化欄位
    for col in ["稽查結果","稽查日期","店名","地址","電話","登錄字號"]:
        if col not in df.columns:
            df[col] = ""

    print(f"\n{'='*50}")
    print(f"欄位：{list(df.columns)}")
    print(f"總筆數：{len(df)}")
    print(f"\n稽查結果分布：")
    print(df["稽查結果"].value_counts().to_string())

    # 各區統計
    summary = (
        df.groupby(["行政區","稽查結果"])
          .size().unstack(fill_value=0).reset_index()
    )
    summary["總筆數"] = summary.iloc[:,1:].sum(axis=1)
    summary = summary.sort_values("總筆數", ascending=False)

    print(f"\n── 台北市各行政區食安不合格紀錄（2026上半年）──")
    print(summary.to_string(index=False))

    # 存檔
    df.to_csv("imap_2026H1_raw.csv", index=False, encoding="utf-8-sig")
    summary.to_csv("imap_2026H1_summary.csv", index=False, encoding="utf-8-sig")
    print(f"\n✅ imap_2026H1_raw.csv")
    print(f"✅ imap_2026H1_summary.csv")
    print(f"\n接著執行：python3 imap_visualize.py")


if __name__ == "__main__":
    main()