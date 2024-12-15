package Config

import (
	"BHLayer2Node/LogWriter"
	"encoding/json"
	"os"
	"strconv"
)

// 定义通信节点的地址配置
type BHNodeAddress struct {
	NodeIPAddress string //节点IP
	NodeGrpcPort  int    //节点grpc端口
	nodeUrl       string //节点访问url
}

// 返回地址字符串
func (b *BHNodeAddress) GetAddrStr() string {
	if b.nodeUrl == "" {
		b.nodeUrl = b.NodeIPAddress + ":" + strconv.Itoa(b.NodeGrpcPort)
	}
	return b.nodeUrl
}

// BHLayer2NodeConfig 定义 Layer2 节点的配置
type BHLayer2NodeConfig struct {
	GrpcPort                   int // gRPC 服务端口
	HttpPort                   int // HTTP 服务端口
	MaxUnprocessedTaskPoolSize int // HTTP 请求池的最大大小
	MaxPendingSchedulePoolSize int
	MaxScheduledTasksPoolSize  int
	MaxCommitSlotItemPoolSize  int
	MaxGrpcRequestPoolSize     int                   // gRPC 请求池的最大大小
	DefaultSlotSize            int                   // 默认的slot大小
	LogPath                    string                // 日志路径
	BHNodeAddressMap           map[int]BHNodeAddress //节点的grpc端口配置，id->nodeIdaddress
	DEBUG                      bool

	// ChainUpper 配置
	FiscoBcosHost         string // FISCO-BCOS 节点地址
	FiscoBcosPort         int    // FISCO-BCOS 节点端口
	GroupID               string // FISCO-BCOS 群组 ID
	PrivateKey            string // 用于签名的私钥
	TLSCaFile             string // TLS CA 证书文件路径
	TLSCertFile           string // TLS 客户端证书路径
	TLSKeyFile            string // TLS 客户端密钥路径
	// ContractAddress       string // 链上合约地址
	QueueBufferSize       int    // 上链队列缓冲区大小
	WorkerCount           int    // Worker 的数量
}

// DefaultBHLayer2NodeConfig 定义默认的配置值
var DefaultBHLayer2NodeConfig = BHLayer2NodeConfig{
	GrpcPort:                   50051, // 默认 gRPC 端口
	HttpPort:                   8080,  // 默认 HTTP 端口
	MaxUnprocessedTaskPoolSize: 100,
	MaxPendingSchedulePoolSize: 100,
	MaxScheduledTasksPoolSize:  100,
	MaxCommitSlotItemPoolSize:  100,
	MaxGrpcRequestPoolSize:     200, // 默认 gRPC 请求池大小
	DefaultSlotSize:            100,
	LogPath:                    "logs/",
	BHNodeAddressMap:           make(map[int]BHNodeAddress, 0),
	DEBUG:                      false,

	// 默认 ChainUpper 配置
	FiscoBcosHost:   "127.0.0.1",
	FiscoBcosPort:   20200,
	GroupID:         "group0",
	PrivateKey:      "145e247e170ba3afd6ae97e88f00dbc976c2345d511b0f6713355d19d8b80b58", 
	TLSCaFile:       "./ChainUpper/ca.crt",
	TLSCertFile:     "./ChainUpper/sdk.crt",
	TLSKeyFile:      "./ChainUpper/sdk.key",
	// ContractAddress: "ChainUpper/contract_address.txt",
	QueueBufferSize: 100000,
	WorkerCount:     3, // 256
}

//var (
//	// GlobalConfig 全局配置实例
//	GlobalConfig *BHLayer2NodeConfig
//	once         sync.Once
//)

// LoadBHLayer2NodeConfig 从指定路径加载配置文件，覆盖默认值
// 如果文件不存在或加载失败，则使用默认配置
func LoadBHLayer2NodeConfig(path string) *BHLayer2NodeConfig {
	//once.Do(func() {
	config := DefaultBHLayer2NodeConfig
	LogWriter.InitGlobalLogWriter(config.LogPath, config.DEBUG)

	// 尝试从配置文件加载
	if path != "" {
		file, err := os.Open(path)
		if err == nil {
			defer file.Close()
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&config)
			if err != nil {
				// 配置文件解析失败时保留默认值
				println("Failed to parse config file, using default values:", err.Error())
			}
		} else {
			// 文件打开失败时保留默认值
			println("Failed to open config file, using default values:", err.Error())
		}
	}

	// 设置全局配置
	return &config

}
