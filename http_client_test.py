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
    send_POST_request(url, request_data)
def oracle_query_epoch():
    # 请求体数据
        request_data = {
            "query": "EvidencePreserveEpochIDQuery",
            "epochID": 8,
        }

        # 发送 POST 请求
        url = "http://127.0.0.1:8080/oracle"
        send_GET_request(url, request_data)

def oracle_query_task():
    request_data = {
        "query": "EvidencePreserveTaskIDQuery",
        "taskID": "SynthTask-0-1738935976",
    }

    # 发送 POST 请求
    url = "http://127.0.0.1:8080/oracle"
    send_GET_request(url, request_data)
def oracle_query_blockchain_latest():
    request_data = {
        "query": "BlockchainLatestInfoQuery"
    }
    url = "http://127.0.0.1:8080/blockchain"
    send_GET_request(url, request_data)
# todo 这里还有blockNumber的Query，暂时手动改，上面的epoch和task也类似
def oracle_query_block():
    request_data = {
#         "query": "BlockchainBlockHashQuery",
#         "blockHash": "0x111"
        "query": "BlockchainBlockNumberQuery",
        "blockNumber": 1,
    }
    url = "http://127.0.0.1:8080/blockchain"
    send_GET_request(url, request_data)
def oracle_query_tx():
    request_data = {
        "query": "BlockchainTransactionQuery",
        "txHash": "0x111"
    }
    url = "http://127.0.0.1:8080/blockchain"
    send_GET_request(url, request_data)
def oracle_query_nodes():
    request_data = {
        "query": "DataSynthMonitorQuery"
    }
    url = "http://127.0.0.1:8080/dataSynth"
    send_GET_request(url, request_data)
def send_POST_request(url, request_data):
    headers = {'Content-Type': 'application/json'}

    response = requests.post(url, json=request_data, headers=headers)
#     print(response.text)
    if response.status_code == 200:
        # 打印响应内容
#         response_data = response.json()
        print(response.json())
#         print(f"Response Status: {response.status_code}")
#         print(f"Response Message: {response_data['msg']}")
#         print(f"Response Code: {response_data['code']}")
    else:
        print(f"Failed to create task. Status code: {response.status_code}")
def send_GET_request(url, request_data):
    headers = {'Content-Type': 'application/json'}

    response = requests.get(url, params=request_data)
    print(response.text)
    if response.status_code == 200:
        # 打印响应内容
    #         response_data = response.json()
        print(response.json())
    #         print(f"Response Status: {response.status_code}")
    #         print(f"Response Message: {response_data['msg']}")
    #         print(f"Response Code: {response_data['code']}")
    else:
        print(f"Failed to query task. Status code: {response.status_code}")
def main():
    print("Welcome to the HTTP Client Shell!")
    print("Type 'create' to create a new task or 'exit' to quit.")

    while True:
        command = input("> ").strip().lower()  # 获取用户输入并转换为小写

        if command == 'create':
            create_task()
        if command == 'epoch':
            oracle_query_epoch()
        if command == 'task':
            oracle_query_task()
        elif command == "bc_latest":
            oracle_query_blockchain_latest()
        elif command == "block":
            oracle_query_block()
        elif command == "tx":
            oracle_query_tx()
        elif command == "node":
            oracle_query_nodes()
        elif command == 'exit':
            print("Exiting the client...")
            break

if __name__ == '__main__':
    main()
