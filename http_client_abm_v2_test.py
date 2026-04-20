import json
from pathlib import Path

import requests


BASE_URL = "http://127.0.0.1:8081"

# 手动测试时优先修改这几个值
DEFAULT_TASK_ID = "TSK-1001"
DEFAULT_STOCK_ID = "600000"
DEFAULT_SECOND_STOCK_ID = "000001"
CACHE_PATH = Path(__file__).resolve().with_name(".abm_v2_http_client_cache.json")


def build_abm_v2_payload():
    return [
        {
            "stockCode": "600000",
            "stockName": "浦发银行",
            "N_FT": 20,
            "N_LMT": 20,
            "N_SMT": 20,
            "N_NT": 20,
            "ALPHA_L": 0.001,
            "ALPHA_S": 0.9,
            "S_FT": 1,
            "fundamental_value": "XX",
            "horizon": "1天 (T+1)",
        },
        {
            "stockCode": "000001",
            "stockName": "平安银行",
            "N_FT": 20,
            "N_LMT": 20,
            "N_SMT": 20,
            "N_NT": 20,
            "ALPHA_L": 0.001,
            "ALPHA_S": 0.9,
            "S_FT": 1,
            "fundamental_value": "XX",
            "horizon": "custom",
            "custom_horizon_date": ["2026-04-14", "2026-04-30"],
        },
    ]


def send_post(path, body, params=None):
    response = requests.post(
        f"{BASE_URL}{path}",
        json=body,
        params=params or {},
        headers={"Content-Type": "application/json"},
        timeout=60,
    )
    print(f"POST {response.url}")
    print(response.text)
    return response


def send_get(path, query=None):
    response = requests.get(
        f"{BASE_URL}{path}",
        params=query or {},
        headers={"Content-Type": "application/json"},
        timeout=60,
    )
    print(f"GET {response.url}")
    print(response.text)
    return response


def create_abm_v2_task():
    response = send_post("/simulation/create-task", build_abm_v2_payload(), params={"isScheduled": "false"})
    if response.status_code != 200:
        print("Create request failed, skip taskId cache update.")
        return

    task_id = extract_task_id_from_create_response(response)
    if task_id:
        save_cached_task_id(task_id)
        print(f"Cached taskId from create response: {task_id}")
        return

    latest_task_id = fetch_latest_platform_task_id()
    if latest_task_id:
        save_cached_task_id(latest_task_id)
        print(f"Cached latest taskId from execution log fallback: {latest_task_id}")
    else:
        print("Create succeeded, but failed to infer latest taskId from execution log.")


def query_abm_parameters():
    send_get("/simulation/abm_parameters")


def query_execution_log():
    send_get("/dashboard/execution_log")


def query_analyzed_stocks(search_type=None, keyword=None):
    query = {}
    if search_type:
        query["searchType"] = search_type
    if keyword:
        query["keyword"] = keyword
    send_get("/market/analyzed_stocks", query)


def build_analytics_query(task_id=None, stock_id=None):
    query = {}
    if task_id:
        query["taskId"] = task_id
    if stock_id:
        query["stockId"] = stock_id
    return query


def query_order_dynamics(task_id=None, stock_id=None, date=None):
    query = build_analytics_query(task_id, stock_id)
    if date:
        query["date"] = date
    send_get("/dashboard/order_dynamics", query)


def query_price_synthesis(task_id=None, stock_id=None):
    send_get("/dashboard/price_synthesis", build_analytics_query(task_id, stock_id))


def query_price_synthesis_download(task_id=None, stock_id=DEFAULT_STOCK_ID):
    task_id = task_id or get_current_task_id()
    response = requests.get(
        f"{BASE_URL}/dashboard/price_synthesis/download",
        params={"taskId": task_id, "stockId": stock_id},
        timeout=60,
    )
    print(f"GET {response.url}")
    print(f"status={response.status_code}")
    print(f"headers={dict(response.headers)}")
    content_type = response.headers.get("Content-Type", "")
    if "text" in content_type or "json" in content_type:
        print(response.text[:1000])
    else:
        print(response.content[:200])


def query_crash_risk(task_id=None, stock_id=None):
    send_get("/dashboard/crash_risk_warning", build_analytics_query(task_id, stock_id))


def query_investor_composition(task_id=None, stock_id=None, date=None, investor_type=None):
    query = build_analytics_query(task_id, stock_id)
    if date:
        query["date"] = date
    if investor_type:
        query["type"] = investor_type
    send_get("/dashboard/investor_composition", query)


def query_performance_comparison(task_id=None, stock_id=None, selected_model=None):
    query = build_analytics_query(task_id, stock_id)
    if selected_model:
        query["selectedModel"] = selected_model
    send_get("/dashboard/performance_comparison", query)


def query_all_analytics(task_id=None, stock_id=None):
    query_order_dynamics(task_id, stock_id)
    query_price_synthesis(task_id, stock_id)
    query_crash_risk(task_id, stock_id)
    query_investor_composition(task_id, stock_id)
    query_performance_comparison(task_id, stock_id)


def query_all_analytics_exact(task_id=None, stock_id=DEFAULT_STOCK_ID):
    query_all_analytics(task_id or get_current_task_id(), stock_id)


def query_all_analytics_by_task(task_id=None):
    query_all_analytics(task_id or get_current_task_id(), None)


def query_all_analytics_latest_by_stock(stock_id=DEFAULT_STOCK_ID):
    query_all_analytics(None, stock_id)


def query_all_analytics_latest_all():
    query_all_analytics(None, None)


def load_cache():
    if not CACHE_PATH.exists():
        return {}
    try:
        return json.loads(CACHE_PATH.read_text(encoding="utf-8"))
    except Exception:
        return {}


def save_cache(cache):
    CACHE_PATH.write_text(json.dumps(cache, ensure_ascii=False, indent=2), encoding="utf-8")


def save_cached_task_id(task_id):
    cache = load_cache()
    cache["task_id"] = task_id
    save_cache(cache)


def get_current_task_id():
    cache = load_cache()
    return cache.get("task_id") or DEFAULT_TASK_ID


def fetch_latest_platform_task_id():
    response = requests.get(f"{BASE_URL}/dashboard/execution_log", timeout=30)
    print(f"GET {response.url}")
    # print(response.text)
    if response.status_code != 200:
        return None
    try:
        payload = response.json()
    except Exception:
        return None

    data = payload.get("data")
    if data is None:
        data = payload.get("Data")
    if isinstance(data, list) and data:
        latest = data[0]
        if isinstance(latest, dict):
            task_id = latest.get("id")
            if isinstance(task_id, str) and task_id:
                return task_id
    return None


def extract_task_id_from_create_response(response):
    try:
        payload = response.json()
    except Exception:
        return None

    data = payload.get("data")
    if isinstance(data, dict):
        task_id = data.get("taskId")
        if isinstance(task_id, str) and task_id:
            return task_id
    return None


def show_current_task_id():
    task_id = get_current_task_id()
    print(f"Current cached taskId: {task_id}")


def print_help():
    print("ABM_V2 HTTP client")
    print("Commands:")
    print("  create              创建 ABM_V2 测试任务")
    print("  abm_params          查询 ABM 参数模板")
    print("  exec_log            查询执行日志")
    print("  analyzed            查询已完成分析股票列表")
    print("  analyzed_code       按股票代码查询已分析股票列表")
    print("  analyzed_name       按股票简称查询已分析股票列表")
    print("  analyzed_task       按任务查询已分析股票列表")
    print("  order               精确查询 600000 订单动态(taskId+stockId)")
    print("  order_task          按任务查询订单动态(仅 taskId)")
    print("  order_latest        查询 600000 最新订单动态(仅 stockId)")
    print("  order_all           查询全部股票最新订单动态(无参数)")
    print("  price               精确查询 600000 价格合成(taskId+stockId)")
    print("  price_task          按任务查询价格合成(仅 taskId)")
    print("  price_latest        查询 600000 最新价格合成(仅 stockId)")
    print("  price_all           查询全部股票最新价格合成(无参数)")
    print("  price_download      下载 600000 合成结果")
    print("  crash               精确查询 600000 崩盘风险(taskId+stockId)")
    print("  crash_task          按任务查询崩盘风险(仅 taskId)")
    print("  crash_latest        查询 600000 最新崩盘风险(仅 stockId)")
    print("  crash_all           查询全部股票最新崩盘风险(无参数)")
    print("  investor            精确查询 600000 投资者构成(taskId+stockId)")
    print("  investor_task       按任务查询投资者构成(仅 taskId)")
    print("  investor_latest     查询 600000 最新投资者构成(仅 stockId)")
    print("  investor_all        查询全部股票最新投资者构成(无参数)")
    print("  investor_custom     查询 600000 投资者构成(type=custom)")
    print("  investor_history    查询 600000 投资者构成(type=history,date=2026-04-19)")
    print("  perf                精确查询 600000 性能对比(taskId+stockId)")
    print("  perf_gbm            精确查询 600000 性能对比(selectedModel=GBM)")
    print("  perf_gan            精确查询 600000 性能对比(selectedModel=GAN)")
    print("  perf_task           按任务查询性能对比(仅 taskId)")
    print("  perf_latest         查询 600000 最新性能对比(仅 stockId)")
    print("  perf_all            查询全部股票最新性能对比(无参数)")
    print("  all                 精确查询 600000 全部分析(taskId+stockId)")
    print("  all_task            按任务查询全部分析(仅 taskId)")
    print("  all_latest          查询 600000 最新全部分析(仅 stockId)")
    print("  all_global          查询全部股票最新全部分析(无参数)")
    print("  order_000001        查询 000001 订单动态")
    print("  all_000001          查询 000001 全部分析")
    print("  current             显示当前缓存的 taskId")
    print("  help                显示帮助")
    print("  exit                退出")
    print("")
    print(f"Default taskId={DEFAULT_TASK_ID}, current taskId={get_current_task_id()}, stockId={DEFAULT_STOCK_ID}")


def main():
    print_help()
    while True:
        command = input("> ").strip().lower()
        if command == "create":
            create_abm_v2_task()
        elif command == "abm_params":
            query_abm_parameters()
        elif command == "exec_log":
            query_execution_log()
        elif command == "analyzed":
            query_analyzed_stocks()
        elif command == "analyzed_code":
            query_analyzed_stocks(search_type="stockCode", keyword=DEFAULT_STOCK_ID)
        elif command == "analyzed_name":
            query_analyzed_stocks(search_type="stockName", keyword="浦发")
        elif command == "analyzed_task":
            query_analyzed_stocks(search_type="task", keyword=get_current_task_id())
        elif command == "order":
            query_order_dynamics(get_current_task_id(), DEFAULT_STOCK_ID)
        elif command == "order_task":
            query_order_dynamics(get_current_task_id(), None)
        elif command == "order_latest":
            query_order_dynamics(None, DEFAULT_STOCK_ID)
        elif command == "order_all":
            query_order_dynamics(None, None)
        elif command == "price":
            query_price_synthesis(get_current_task_id(), DEFAULT_STOCK_ID)
        elif command == "price_task":
            query_price_synthesis(get_current_task_id(), None)
        elif command == "price_latest":
            query_price_synthesis(None, DEFAULT_STOCK_ID)
        elif command == "price_all":
            query_price_synthesis(None, None)
        elif command == "price_download":
            query_price_synthesis_download()
        elif command == "crash":
            query_crash_risk(get_current_task_id(), DEFAULT_STOCK_ID)
        elif command == "crash_task":
            query_crash_risk(get_current_task_id(), None)
        elif command == "crash_latest":
            query_crash_risk(None, DEFAULT_STOCK_ID)
        elif command == "crash_all":
            query_crash_risk(None, None)
        elif command == "investor":
            query_investor_composition(get_current_task_id(), DEFAULT_STOCK_ID)
        elif command == "investor_task":
            query_investor_composition(get_current_task_id(), None)
        elif command == "investor_latest":
            query_investor_composition(None, DEFAULT_STOCK_ID)
        elif command == "investor_all":
            query_investor_composition(None, None)
        elif command == "investor_custom":
            query_investor_composition(get_current_task_id(), DEFAULT_STOCK_ID, investor_type="custom")
        elif command == "investor_history":
            query_investor_composition(get_current_task_id(), DEFAULT_STOCK_ID, date="2026-04-19", investor_type="history")
        elif command == "perf":
            query_performance_comparison(get_current_task_id(), DEFAULT_STOCK_ID)
        elif command == "perf_gbm":
            query_performance_comparison(get_current_task_id(), DEFAULT_STOCK_ID, selected_model="GBM")
        elif command == "perf_gan":
            query_performance_comparison(get_current_task_id(), DEFAULT_STOCK_ID, selected_model="GAN")
        elif command == "perf_task":
            query_performance_comparison(get_current_task_id(), None)
        elif command == "perf_latest":
            query_performance_comparison(None, DEFAULT_STOCK_ID)
        elif command == "perf_all":
            query_performance_comparison(None, None)
        elif command == "all":
            query_all_analytics_exact()
        elif command == "all_task":
            query_all_analytics_by_task()
        elif command == "all_latest":
            query_all_analytics_latest_by_stock()
        elif command == "all_global":
            query_all_analytics_latest_all()
        elif command == "order_000001":
            query_order_dynamics(get_current_task_id(), DEFAULT_SECOND_STOCK_ID)
        elif command == "all_000001":
            query_all_analytics_exact(stock_id=DEFAULT_SECOND_STOCK_ID)
        elif command == "current":
            show_current_task_id()
        elif command == "help":
            print_help()
        elif command == "exit":
            print("Exiting the client...")
            break
        else:
            print("Unknown command. Type 'help' to list commands.")


if __name__ == "__main__":
    main()
