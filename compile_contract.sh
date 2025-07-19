#!/usr/bin/env bash
# -----------------------------------------------------------------------------
# compile_contract.sh
#
# 用途：
#   在 FISCO-BCOS Go-SDK 环境下，自动化地编译指定的 Solidity 合约，
#   并生成对应的 .bin、.abi 以及 Go 绑定文件 (.go)。
#
# 使用方法：
#   将脚本置于项目根目录（如 /root/rappa/RappaMaster/），然后：
#
#     chmod +x compile_contract.sh
#     ./compile_contract.sh <合约文件路径> <go-sdk 根目录>
#
#   例如：
#     ./compile_contract.sh \
#       ChainUpper/contract/storeData/StoreData.sol \
#       /root/go-sdk
#
# 参数说明：
#   $1 — 合约文件的相对或绝对路径（.sol 文件）
#   $2 — Go-SDK 根目录（包含 solc-0.8.11 和 abigen 二进制文件）
#
# 脚本功能：
#   1. 解析并创建一个与合约同名的子目录于 Go-SDK 根目录下
#   2. 复制 .sol 合约到该子目录
#   3. 调用 solc-0.8.11 生成 .bin 与 .abi 文件
#   4. 调用 abigen 生成 Go 绑定 (.go)
#   5. 将生成的 .bin、.abi、.go 文件复制回合约原目录
#
# 要求：
#   - Go-SDK 目录下必须存在可执行文件 solc-0.8.11 与 abigen
#   - solc-0.8.11 执行/go-sdk/tools/download_solc.sh安装
#
# -----------------------------------------------------------------------------

set -euo pipefail

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 参数检查
if [ "$#" -ne 2 ]; then
  echo "Usage: $0 <contract_file_path> <go_sdk_path>"
  exit 1
fi

# 将传入路径解析为绝对路径，支持相对调用
CONTRACT_FILE="$(realpath "$1")"
GO_SDK_PATH="$(realpath "$2")"

# derive names and paths
CONTRACT_DIR="$(dirname "$CONTRACT_FILE")"
CONTRACT_FILENAME="$(basename "$CONTRACT_FILE")"
CONTRACT_NAME="${CONTRACT_FILENAME%.sol}"
DEST_DIR="$GO_SDK_PATH/$CONTRACT_NAME"

# 1. 在 go-sdk 目录下准备目录并复制合约
mkdir -p "$DEST_DIR"
cp "$CONTRACT_FILE" "$DEST_DIR/"

# 进入 go-sdk 根目录
cd "$GO_SDK_PATH"

# 2. 生成 bin 和 abi
if [ ! -x "./solc-0.8.11" ]; then
  echo "Error: solc-0.8.11 not found or not executable in $GO_SDK_PATH"
  exit 1
fi
./solc-0.8.11 --bin --abi -o "./$CONTRACT_NAME" "./$CONTRACT_NAME/$CONTRACT_FILENAME"

# 3. 生成 Go 绑定文件
if [ ! -x "./abigen" ]; then
  echo "Error: abigen not found or not executable in $GO_SDK_PATH"
  exit 1
fi
./abigen \
  --bin   "./$CONTRACT_NAME/$CONTRACT_NAME.bin" \
  --abi   "./$CONTRACT_NAME/$CONTRACT_NAME.abi" \
  --pkg   "$CONTRACT_NAME" \
  --type  "$CONTRACT_NAME" \
  --out   "./$CONTRACT_NAME/$CONTRACT_NAME.go"

# 4. 将生成结果拷贝回项目合约目录
cp -r "$DEST_DIR/"* "$CONTRACT_DIR/"

echo "✅ Contract '$CONTRACT_NAME' compiled successfully.
  → ABI:     $CONTRACT_DIR/$CONTRACT_NAME.abi
  → BIN:     $CONTRACT_DIR/$CONTRACT_NAME.bin
  → Go file: $CONTRACT_DIR/$CONTRACT_NAME.go"
