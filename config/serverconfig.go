package config

import (
	"github.com/go-ini/ini"
	"log"
	"strings"
)

// 服务端配置
type ServerConfig struct {
	Port          uint32 // 服务端口
	Key           string // 6-16 个字符，用于身份校验
	MinAccessPort uint32 // 最小访问端口，最小值 1024
	MaxAccessPort uint32 // 最大访问端口，最大值 65535
}

// 检查端口是否在允许范围内，不含边界
func (c *ServerConfig) PortInRange(port uint32) bool {
	return port > c.MinAccessPort && port < c.MaxAccessPort
}

// 从参数中解析配置
func _parseServerConfig(args []string) ServerConfig {
	if len(args) < 2 {
		log.Fatalln("More args in need")
	}
	// 0 key
	key := strings.TrimSpace(args[0])

	// 1 port
	port, err := parsePort(args[1])
	if err != nil || !checkPort(port) {
		log.Fatalln("Fail to parse args.", args)
	}

	// 2 access port range
	portRange := strings.Split(args[2], "-")
	if len(portRange) != 2 {
		log.Fatalln("Fail to parse args.", args)
	}

	minAccessPort, err := parsePort(portRange[0])
	if err != nil || !checkPort(minAccessPort) {
		log.Fatalln("Fail to parse args.", args)
	}
	maxAccessPort, err := parsePort(portRange[1])
	if err != nil || !checkPort(maxAccessPort) {
		log.Fatalln("Fail to parse args.", args)
	}
	// 检查范围是否正确，确保范围内至少有一个元素
	if maxAccessPort-minAccessPort < 2 {
		log.Fatalln("Fail to parse args.", args)
	}

	return ServerConfig{
		Port:          port,
		Key:           key,
		MinAccessPort: minAccessPort,
		MaxAccessPort: maxAccessPort,
	}
}

// 从配置文件中加载配置
func _loadServerConfig() ServerConfig {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalln("Fail to load config.ini", err.Error())
	}
	server := func(key string) *ini.Key {
		return cfg.Section("server").Key(key)
	}

	args := make([]string, 3)
	args[0] = server("key").String()
	args[1] = server("port").String()
	args[2] = server("access-port-range").String()

	return _parseServerConfig(args)
}

// 初始化服务端配置，支持从参数中读取或者从配置文件中读取
func InitServerConfig(args []string) ServerConfig {
	if len(args) == 0 {
		return _loadServerConfig()
	}
	return _parseServerConfig(args)
}
