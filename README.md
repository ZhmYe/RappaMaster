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

      

