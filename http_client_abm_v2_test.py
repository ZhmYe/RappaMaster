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

    latest_task_id = fetch_latest_platform_task_id()
    if latest_task_id:
        save_cached_task_id(latest_task_id)
        print(f"Cached latest taskId: {latest_task_id}")
    else:
        print("Create succeeded, but failed to infer latest taskId from execution log.")


def query_abm_parameters():
    send_get("/simulation/abm-parameters")


def query_execution_log():
    send_get("/dashboard/execution_log")


def query_analyzed_stocks():
    send_get("/market/analyzed-stocks")


def query_order_dynamics(task_id=None, stock_id=DEFAULT_STOCK_ID):
    task_id = task_id or get_current_task_id()
    send_get("/dashboard/order_dynamics", {"taskId": task_id, "stockId": stock_id})


def query_price_synthesis(task_id=None, stock_id=DEFAULT_STOCK_ID):
    task_id = task_id or get_current_task_id()
    send_get("/dashboard/price_synthesis", {"taskId": task_id, "stockId": stock_id})


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


def query_crash_risk(task_id=None, stock_id=DEFAULT_STOCK_ID):
    task_id = task_id or get_current_task_id()
    send_get("/dashboard/crash_risk_warning", {"taskId": task_id, "stockId": stock_id})


def query_investor_composition(task_id=None, stock_id=DEFAULT_STOCK_ID):
    task_id = task_id or get_current_task_id()
    send_get("/dashboard/investor_composition", {"taskId": task_id, "stockId": stock_id})


def query_performance_comparison(task_id=None, stock_id=DEFAULT_STOCK_ID):
    task_id = task_id or get_current_task_id()
    send_get("/dashboard/performance_comparison", {"taskId": task_id, "stockId": stock_id})


def query_all_analytics(task_id=None, stock_id=DEFAULT_STOCK_ID):
    task_id = task_id or get_current_task_id()
    query_order_dynamics(task_id, stock_id)
    query_price_synthesis(task_id, stock_id)
    query_crash_risk(task_id, stock_id)
    query_investor_composition(task_id, stock_id)
    query_performance_comparison(task_id, stock_id)


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
    print("  order               查询 600000 订单动态")
    print("  price               查询 600000 价格合成")
    print("  price_download      下载 600000 合成结果")
    print("  crash               查询 600000 崩盘风险")
    print("  investor            查询 600000 投资者构成")
    print("  perf                查询 600000 性能对比")
    print("  all                 查询 600000 全部分析")
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
        elif command == "order":
            query_order_dynamics()
        elif command == "price":
            query_price_synthesis()
        elif command == "price_download":
            query_price_synthesis_download()
        elif command == "crash":
            query_crash_risk()
        elif command == "investor":
            query_investor_composition()
        elif command == "perf":
            query_performance_comparison()
        elif command == "all":
            query_all_analytics()
        elif command == "order_000001":
            query_order_dynamics(stock_id=DEFAULT_SECOND_STOCK_ID)
        elif command == "all_000001":
            query_all_analytics(stock_id=DEFAULT_SECOND_STOCK_ID)
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
