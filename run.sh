export LD_LIBRARY_PATH=/usr/local/lib
go run main.go &

MASTER_PID=$!

# 将 PID 写入 master.pid 文件
echo $MASTER_PID > master.pid
echo "RappaMaster started with PID: $MASTER_PID"

wait $MASTER_PID
