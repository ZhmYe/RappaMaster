package Config

import "BHLayer2node/LogWriter"

// BHLayer2NodeConfig 定义 Layer2 节点的配置
type BHLayer2NodeConfig struct {
	GrpcPort               int    // gRPC 服务端口
	HttpPort               int    // HTTP 服务端口
	MaxHttpRequestPoolSize int    // HTTP 请求池的最大大小
	MaxGrpcRequestPoolSize int    // gRPC 请求池的最大大小
	DefaultSlotSize        int    // 默认的slot大小
	LogPath                string // 日志路径
	DEBUG                  bool
}

// DefaultBHLayer2NodeConfig 定义默认的配置值
var DefaultBHLayer2NodeConfig = BHLayer2NodeConfig{
	GrpcPort:               50051, // 默认 gRPC 端口
	HttpPort:               8080,  // 默认 HTTP 端口
	MaxHttpRequestPoolSize: 100,   // 默认 HTTP 请求池大小
	MaxGrpcRequestPoolSize: 200,   // 默认 gRPC 请求池大小
	DefaultSlotSize:        100,
	LogPath:                "logs/",
	DEBUG:                  false,
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
	//// 尝试从配置文件加载
	//if path != "" {
	//	file, err := os.Open(path)
	//	if err == nil {
	//		defer file.Close()
	//		decoder := json.NewDecoder(file)
	//		err = decoder.Decode(&config)
	//		if err != nil {
	//			// 配置文件解析失败时保留默认值
	//			println("Failed to parse config file, using default values:", err.Error())
	//		}
	//	} else {
	//		// 文件打开失败时保留默认值
	//		println("Failed to open config file, using default values:", err.Error())
	//	}
	//}

	// 设置全局配置
	return &config
	//})
}
