package test

import (
	"chuantou/config"
	"chuantou/core"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"log"
	"testing"
	"time"
)

// -----------------------------------------------------------------------------

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

// netbus
func TestServer(t *testing.T) {
	cfg := config.ServerConfig{
		Key:           "winshu",
		Port:          6666,
		MinAccessPort: 10000,
		MaxAccessPort: 20000,
	}
	core.Server(cfg)
}

func TestClient(t *testing.T) {
	cfg := config.ClientConfig{
		Key: "winshu",
		ServerAddr: config.NetAddress{
			IP: "127.0.0.1", Port: 6666,
		},
		LocalAddr: []config.NetAddress{
			{"127.0.0.1", 3306, 13306},
		},
		TunnelCount: 1,
	}
	core.Client(cfg)
}

func TestParseNetAddress(t *testing.T) {
	addr, _ := config.ParseNetAddress("127.0.0.1:3389:13389")
	println(addr.String())
	println(addr.FullString())
}

func TestProtocol(t *testing.T) {
	seed := "winshu"
	key, _ := config.NewKey(seed, "2019-12-31")
	fmt.Println(key)
	fmt.Println(config.CheckKey(seed, key))
}

func TestGetMachineID(t *testing.T) {
	fmt.Println(machineid.ID())
}

func TestClient2(t *testing.T) {
	cfg := config.ClientConfig{
		Key: "winshu",
		ServerAddr: config.NetAddress{
			IP: "127.0.0.1", Port: 6666,
		},
		LocalAddr: []config.NetAddress{
			{"127.0.0.1", 3306, 13306},
		},
		TunnelCount: 1,
	}
	core.Client(cfg)
}

func TestTick(t *testing.T) {
	c := time.Tick(5 * time.Second)
	for {
		fmt.Println(c)
		<-c
	}
}
