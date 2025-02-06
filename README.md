## RappaMaster

### 一、环境准备

- FISCO-BCOS

    - 搭建fiscobcos3.0版本：参考 [https://fisco-bcos-documentation.readthedocs.io/zh-cn/latest/docs/installation.html#id2](https://fisco-bcos-doc.readthedocs.io/zh-cn/latest/docs/quick_start/air_installation.html)
    - 配置文件：将fisco目录下 127.0.0.1/sdk文件夹下的ca.crt、sdk.crt、sdk.key替换到文件夹相应位置
    - 启动FISCO-BCOS：运行 fisco/nodes/127.0.0.1/start_all.sh

- 异步上链

    - FISCO-BCOS GO-SDK

        - go-sdk需要依赖csdk的动态库

          ```sh
          # 下面的脚本帮助用户下载bcos-c-sdk的动态库到/usr/local/lib目录下
          ./tools/download_csdk_lib.sh
          
          # 下载完成时需要设置动态库的搜索路径
          export LD_LIBRARY_PATH=${PWD}/lib # 上面设置到了/usr/local/lib下因此是export LD_LIBRARY_PATH=/usr/local/lib
          ```

### 二、HTTP测试客户端 (http_client_test.py)

- 功能概述：这个脚本提供了一个简单的 HTTP 客户端，能够发送请求与`Master`中的`HttpEngine`交互。它通过 `requests` 库模拟与服务端的 HTTP 通信，通过shell的形式提供交互，目前提供以下命令

    - `create`:创建一个合成任务

        - 请求方式: POST

        - URL: `http://127.0.0.1:8080/create`

        - 请求体示例

          ```json
          {
              "model": "CTGAN",
              "params": {
                  "condition_column": "native-country",
                  "condition_value": "United-States"
              },
              "size": 50,
              "isReliable": true
          }
          
          ```

        - 会返回一个`taskID`，可修改代码中对`task`的请求

    - `task`：查询一个合成任务

        - 请求方式：GET

        - URL: `http://127.0.0.1:8080/oracle`

        - 请求体示例

          ```json
          {
              "query": "EvidencePreserveEpochIDQuery",
              "epochID": 8
          }
          ```

        - 会返回一个合成任务在`oracle`中的记录细节

    - `epoch`：查询一个`epoch`

        - 请求方式: `GET`

        - URL: `http://127.0.0.1:8080/oracle`

        - 请求体示例

          ```json
          {
              "query": "EvidencePreserveEpochIDQuery",
              "epochID": 8
          }
          ```

        - 会返回一个epoch在`oracle`的记录细节

### 三、运行方式

- 先运行`Executor`

  ```shell
  # 在对应路径下
  python3 main.py
  # 等待所有模型加载完全
  ```

- 运行`Master`

  ```shell
  # 在对应路径下
  bash run.sh
  ```

- 运行HTTP客户端

  ```shell
  python3 http_client_test.py
  
  # 在shell中输入
  create # 新建一个合成任务，这样Master才会开始调度
  
  
  # 可选，运行过程中或运行完成后
  epoch # 查询epoch
  task # 查询task
  ```

  