"""
iMAP API 偵測腳本
用來找出實際回傳資料的 API 端點與解析方式
"""

import requests
import json
import urllib3
urllib3.disable_warnings()

BASE_URL = "https://imap.health.gov.tw"

HEADERS = {
    "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
    "Accept": "*/*",
    "Accept-Language": "zh-TW,zh;q=0.9,en;q=0.8",
    "Referer": "https://imap.health.gov.tw/App_Prog/SubjectFoodSafetyInfo.aspx",
    "X-Requested-With": "XMLHttpRequest",
}

session = requests.Session()
session.headers.update(HEADERS)
session.verify = False

# ── 已知的行政區代碼（從你的輸出看到的）──
DISTRICT_CODES = {
    "松山區": "0101",
    "信義區": "0117",
    "大安區": "0102",
    "內湖區": "0110",
    "士林區": "0118",
    "中山區": "0109",
    "北投區": "0119",
    "文山區": "0120",
    "南港區": "0112",
    "萬華區": "0111",
    "中正區": "0115",
    "大同區": "0116",
}

# ── 測試幾個已知的 AJAX endpoint 模式 ──
POSSIBLE_ENDPOINTS = [
    "/App_Prog/SubjectFoodSafetyInfo.aspx/GetData",
    "/App_Prog/Handler/FoodSafetyHandler.ashx",
    "/App_Prog/FoodSafetyData.aspx",
    "/App_Prog/SubjectFoodSafetyInfo.aspx",
    "/WebService/FoodSafetyService.asmx",
    "/api/FoodSafety",
]

def test_endpoints():
    print("=" * 60)
    print("【測試已知 AJAX 端點】")
    print("=" * 60)
    
    test_payload = {
        "district": "0101",
        "sdate": "2026/01/01",
        "edate": "2026/06/30",
        "type": "food",
    }
    
    for ep in POSSIBLE_ENDPOINTS:
        url = BASE_URL + ep
        try:
            r = session.post(url, data=test_payload, timeout=10)
            print(f"\n{url}")
            print(f"  Status: {r.status_code} | Content-Type: {r.headers.get('Content-Type','?')}")
            print(f"  Response (前200字): {r.text[:200]!r}")
        except Exception as e:
            print(f"\n{url} → Error: {e}")


def inspect_main_page():
    """取得主頁原始 HTML，尋找 JS 裡的 API 呼叫"""
    print("\n" + "=" * 60)
    print("【分析主頁 JavaScript，尋找 API 端點】")
    print("=" * 60)
    
    url = BASE_URL + "/App_Prog/SubjectFoodSafetyInfo.aspx"
    r = session.get(url, timeout=20)
    
    import re
    # 找 ajax / fetch / $.post / $.get / xmlhttp
    patterns = [
        r"url\s*[:=]\s*['\"]([^'\"]+)['\"]",
        r"\.ajax\s*\(\s*['\"]([^'\"]+)['\"]",
        r"\.post\s*\(\s*['\"]([^'\"]+)['\"]",
        r"\.get\s*\(\s*['\"]([^'\"]+)['\"]",
        r"fetch\s*\(\s*['\"]([^'\"]+)['\"]",
        r"XMLHttpRequest.*open.*['\"]([^'\"]+)['\"]",
        r"ashx[^'\"]*",
        r"\.asmx[^'\"]*",
        r"/api/[^\s'\"]+",
        r"Handler[^\s'\"]+",
    ]
    
    found = set()
    for pat in patterns:
        for m in re.findall(pat, r.text, re.IGNORECASE):
            if len(m) > 3:
                found.add(m)
    
    print("\n找到的可能 URL/端點：")
    for f in sorted(found):
        print(f"  {f}")
    
    # 也印出所有外部 JS 檔案
    from bs4 import BeautifulSoup
    soup = BeautifulSoup(r.text, "html.parser")
    print("\n載入的 JS 檔案：")
    for script in soup.find_all("script", src=True):
        print(f"  {script['src']}")
    
    # 找所有 form action
    print("\nForm actions：")
    for form in soup.find_all("form"):
        print(f"  action={form.get('action','')!r}  method={form.get('method','')!r}")
    
    # 存原始 HTML 供手動檢查
    with open("imap_mainpage.html", "w", encoding="utf-8") as f:
        f.write(r.text)
    print("\n✅ 主頁 HTML 已存為 imap_mainpage.html（可用瀏覽器開啟或搜尋關鍵字）")


def try_direct_query():
    """
    嘗試直接打 POST，印出完整 raw response。
    參數名稱從你現有的腳本沿用。
    """
    print("\n" + "=" * 60)
    print("【嘗試直接 POST 查詢，印出原始回應】")
    print("=" * 60)
    
    url = BASE_URL + "/App_Prog/SubjectFoodSafetyInfo.aspx"
    
    # 先 GET 一次取 cookies + viewstate
    r0 = session.get(url, timeout=20)
    from bs4 import BeautifulSoup
    soup = BeautifulSoup(r0.text, "html.parser")
    
    vs = soup.find("input", {"id": "__VIEWSTATE"})
    vsg = soup.find("input", {"id": "__VIEWSTATEGENERATOR"})
    ev = soup.find("input", {"id": "__EVENTVALIDATION"})
    
    # 列出所有 input 欄位（找到真實欄位名稱）
    print("\n【所有 <input> 欄位】")
    for inp in soup.find_all("input"):
        n = inp.get("name", "")
        t = inp.get("type", "text")
        v = (inp.get("value") or "")[:50]
        if n:
            print(f"  name={n!r:60s} type={t:10s} value={v!r}")
    
    print("\n【所有 <select> 下拉選單】")
    for sel in soup.find_all("select"):
        n = sel.get("name","")
        opts = [(o.get("value",""), o.get_text(strip=True)) for o in sel.find_all("option")]
        print(f"\n  name={n!r}")
        for val, text in opts[:20]:
            print(f"    value={val!r}  text={text!r}")
    
    print("\n\n存 HTML 到 imap_mainpage.html 供進一步分析")
    with open("imap_mainpage.html", "w", encoding="utf-8") as f:
        f.write(r0.text)


if __name__ == "__main__":
    inspect_main_page()
    try_direct_query()
    test_endpoints()
    
    print("\n" + "=" * 60)
    print("📋 下一步：")
    print("  1. 開啟 Chrome DevTools (F12) → Network 頁籤")
    print("  2. 開啟 https://imap.health.gov.tw/App_Prog/SubjectFoodSafetyInfo.aspx")
    print("  3. 在網頁上選 臺北市 → 松山區 → 按查詢")
    print("  4. 在 Network 裡找出 XHR/Fetch 請求")
    print("  5. 右鍵 → Copy → Copy as cURL，貼給 Claude")
    print("=" * 60)