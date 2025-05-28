import requests
import time
import csv
from requests.exceptions import RequestException

# 配置参数
BASE_URL = "http://127.0.0.1:8081/collect"  # 根据实际服务地址修改
TASK_IDS = [
    "SynthTask-10-1744884851",
    "SynthTask-11-1744884891",
    "SynthTask-12-1744884941",
    "SynthTask-13-1744885001",
    "SynthTask-14-1744885071",
    "SynthTask-15-1744885151",
    "SynthTask-16-1744885241",
    "SynthTask-17-1744885342",
    "SynthTask-18-1744885452",
    "SynthTask-19-1744885572",
    "SynthTask-20-1744886387",
    "SynthTask-21-1744886477",
    "SynthTask-22-1744886577",
    "SynthTask-23-1744886687",
    "SynthTask-24-1744886807"
]
PARAMS_TEMPLATE = {
    "query": "CollectTaskQuery",
}
TIMEOUT = 30  # 请求超时时间（秒）

def main():
    with open("request_logs.csv", "w", newline="", encoding="utf-8") as csvfile:
        csv_writer = csv.writer(csvfile)
        csv_writer.writerow(["TaskID", "Duration(s)", "HTTP Status", "Success", "Error"])
        s=1000
        for task_id in TASK_IDS:
            params = {**PARAMS_TEMPLATE, "taskID": task_id,"size":s}
            s+=1000
            start_time = time.time()
            success = False
            status_code = None
            error_msg = ""

            try:
                response = requests.get(
                    BASE_URL,
                    params=params,
                    timeout=TIMEOUT,
                    stream=True  # 流式接收避免立即下载大文件
                )
                response.raise_for_status()  # 检查HTTP错误

                # 如果需要实际下载文件可以在此处理响应内容
                # content = response.content

                status_code = response.status_code
                success = True
            except RequestException as e:
                error_msg = str(e)
                if hasattr(e, "response") and e.response is not None:
                    status_code = e.response.status_code
            finally:
                duration = time.time() - start_time
                print(
                    f"Task {task_id} | "
                    f"Time: {duration:.2f}s | "
                    f"Status: {status_code or '---'} | "
                    f"Success: {success} | "
                    f"Error: {error_msg[:50]}"
                )
                csv_writer.writerow([
                    task_id,
                    f"{duration:.2f}",
                    status_code or "N/A",
                    success,
                    error_msg[:500]  # 限制错误信息长度
                ])

if __name__ == "__main__":
    main()