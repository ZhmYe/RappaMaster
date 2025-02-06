import requests
import json

def create_task():
    # 请求体数据
    request_data = {
        "model": "CTGAN",
        "params": {
            "condition_column": "native-country",
            "condition_value": "United-States"
        },
        "size": 50,
        "isReliable": True
    }

    # 发送 POST 请求
    url = "http://127.0.0.1:8080/create"  # 修改为你的实际服务器地址和端口
    headers = {'Content-Type': 'application/json'}

    response = requests.post(url, json=request_data, headers=headers)
    print(response.text)
    if response.status_code == 200:
        # 打印响应内容
        response_data = response.json()
        print(f"Response Status: {response.status_code}")
        print(f"Response Message: {response_data['message']}")
        print(f"Response Code: {response_data['code']}")
    else:
        print(f"Failed to create task. Status code: {response.status_code}")

def main():
    print("Welcome to the HTTP Client Shell!")
    print("Type 'create' to create a new task or 'exit' to quit.")

    while True:
        command = input("> ").strip().lower()  # 获取用户输入并转换为小写

        if command == 'create':
            create_task()
        elif command == 'exit':
            print("Exiting the client...")
            break
        else:
            print("Unknown command:", command)
            print("Valid commands: create, exit")

if __name__ == '__main__':
    main()
