package config

type FBConfig struct {
	// ChainUpper 配置
	FiscoBcosHost string // FISCO-BCOS node address
	FiscoBcosPort int    // FISCO-BCOS node ip
	GroupID       string // FISCO-BCOS group id
	PrivateKey    string // private key
	TLSCaFile     string // TLS CA path
	TLSCertFile   string // TLS client cert path
	TLSKeyFile    string // TLS key path
}

func (fc *FBConfig) SetDefault() {
	fc.FiscoBcosHost = "127.0.0.1"
	fc.FiscoBcosPort = 20200
	fc.GroupID = "group0"
	fc.PrivateKey = "145e247e170ba3afd6ae97e88f00dbc976c2345d511b0f6713355d19d8b80b58"
	fc.TLSCaFile = "./keys/ca.crt"
	fc.TLSCertFile = "./keys/sdk.crt"
	fc.TLSKeyFile = "./keys/sdk.key"
}
