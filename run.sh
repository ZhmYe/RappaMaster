export LD_LIBRARY_PATH=/usr/local/lib

export CGO_LDFLAGS="-L/usr/local/lib"
export CGO_CFLAGS="-I/path/to/bcos-c-sdk/include"

go run main.go &

MASTER_PID=$!

# 将 PID 写入 master.pid 文件
echo $MASTER_PID > master.pid
echo "RappaMaster started with PID: $MASTER_PID"

wait $MASTER_PID
