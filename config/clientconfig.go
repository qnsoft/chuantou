package config

import (
	"encoding/json"
	"fmt"
	"github.com/go-ini/ini"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	// 默认最大隧道数
	MinTunnelCount = 1
	MaxTunnelCount = 5
)

// 客户端配置
type ClientConfig struct {
	Key         string       // 参考服务端配置 custom-port-key random-port-key
	ServerAddr  NetAddress   // 服务端地址
	LocalAddr   []NetAddress // 内网服务地址及映射端口
	TunnelCount int          // 隧道条数(1-5)
}

func (p *ClientConfig) Local(port uint32) NetAddress {
	for index := range p.LocalAddr {
		if p.LocalAddr[index].Port2 == port {
			return p.LocalAddr[index]
		}
	}
	return NetAddress{}
}

// 从参数中解析配置
func _parseClientConfig(args []string) ClientConfig {
	if len(args) < 3 {
		log.Fatalln("More args in need.", args)
	}

	config := ClientConfig{TunnelCount: MinTunnelCount}
	var ok bool

	// 1 Key
	config.Key = strings.TrimSpace(args[0])
	// 2 ServerAddr
	if config.ServerAddr, ok = ParseNetAddress(strings.TrimSpace(args[1])); !ok {
		log.Fatalln("Fail to parse ServerAddr")
	}
	// 3 LocalAddr
	if config.LocalAddr, ok = ParseNetAddresses(strings.TrimSpace(args[2])); !ok {
		log.Fatalln("Fail to parse LocalAddr")
	}
	// 4 TunnelCount
	if len(args) >= 4 {
		var err error
		if config.TunnelCount, err = strconv.Atoi(args[3]); err != nil {
			log.Fatalln("Fail to parse TunnelCount")
		}
		if config.TunnelCount > MaxTunnelCount {
			config.TunnelCount = MaxTunnelCount
		}
		if config.TunnelCount < MinTunnelCount {
			config.TunnelCount = MinTunnelCount
		}
	}
	return config
}

// 从配置文件中加载配置
func _loadClientConfig() ClientConfig {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalln("Fail to load config.ini", err.Error())
	}
	client := func(key string) *ini.Key {
		return cfg.Section("client").Key(key)
	}
	args := make([]string, 4)
	args[0] = client("key").String()
	addr, _ := net.ResolveIPAddr("ip", client("server-host").String()[0:strings.Index(client("server-host").String(), ":")])
	args[1] = addr.IP.String() + client("server-host").String()[strings.Index(client("server-host").String(), ":"):]
	var portList []string
	json.Unmarshal([]byte(client("local-host-mapping").String()), &portList)
	str_mapping := ""
	for _, _port := range portList {
		if len(str_mapping) > 0 {
			str_mapping += ","
		}
		str_mapping += fmt.Sprintf("%s", _port)
	}
	args[2] = str_mapping
	args[3] = client("tunnel-count").String()
	return _parseClientConfig(args)
}

// 初始化客户端配置，支持从参数中读取或者从配置文件中读取
func InitClientConfig(args []string) ClientConfig {
	if len(args) == 0 {
		return _loadClientConfig()
	}
	return _parseClientConfig(args)
}
