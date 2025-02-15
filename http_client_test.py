import requests
import json

def create_task():
    # 请求体数据
    request_data = {
        "model": "FINKAN",
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
        "taskID": "SynthTask-0-1739524707",
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
        "query": "BlockchainBlockHashQuery",
        "blockHash": "0x1ca92b9f55a9f977f85f7d4a0f07c31ba2b8e75a903d7e0fe0999e15a351b19c",
        # "query": "BlockchainBlockNumberQuery",
        # "blockNumber": 88,
    }
    url = "http://127.0.0.1:8080/blockchain"
    send_GET_request(url, request_data)
def oracle_query_tx():
    request_data = {
        "query": "BlockchainTransactionQuery",
        "txHash": "0x5e52a917657f32bf0fe8b894c4bd2a4e8410a50b1237d487c0ddaf9ddde622bd"
    }
    url = "http://127.0.0.1:8080/blockchain"
    send_GET_request(url, request_data)
def oracle_query_nodes():
    request_data = {
        "query": "NodesStatusQuery"
    }
    url = "http://127.0.0.1:8080/dataSynth"
    send_GET_request(url, request_data)
def oracle_query_date_synth():
    request_data = {
        "query": "DateSynthDataQuery"
    }
    url = "http://127.0.0.1:8080/dataSynth"
    send_GET_request(url, request_data)
def oracle_query_date_tx():
    request_data = {
        "query": "DateTransactionQuery"
    }
    url = "http://127.0.0.1:8080/dataSynth"
    send_GET_request(url, request_data)
def oracle_query_tasks():
    request_data = {
        "query": "SynthTaskQuery"
    }
    url = "http://127.0.0.1:8080/oracle"
    send_GET_request(url, request_data)
def oracle_query_collect():
    request_data = {
        "query": "CollectTaskQuery",
        "taskID": "SynthTask-0-1739610529",
        "size": 50
    }
    url = "http://127.0.0.1:8080/collect"
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
        print(command)
        if command == 'create':
            create_task()
        if command == 'epoch':
            oracle_query_epoch()
        if command == 'task':
            oracle_query_task()
        if command == "bc_latest":
            oracle_query_blockchain_latest()
        if command == "block":
            oracle_query_block()
        if command == "tx":
            oracle_query_tx()
        if command == "node":
            oracle_query_nodes()
        if command == "date_synth":
            oracle_query_date_synth()
        if command == "date_tx":
            oracle_query_date_tx()
        if command == "tasks":
            oracle_query_tasks()
        if command == "collect":
            oracle_query_collect()
        if command == 'exit':
            print("Exiting the client...")
            break

if __name__ == '__main__':
    main()
