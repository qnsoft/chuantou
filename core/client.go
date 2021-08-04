package core

import (
	"chuantou/config"
	"github.com/denisbrodbeck/machineid"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
)

// 客户端ID，启动时就确定了
var clientID string

func init() {
	if id, err := machineid.ID(); err == nil {
		clientID = strings.ReplaceAll(id, "-", "")
		log.Println("Get client machine ID :", clientID)
	} else {
		log.Println("Fail to get machine ID")
		os.Exit(0)
	}
}

// 处理客户端连接
func handleClientConnection(cfg config.ClientConfig, index int) {
	connChan := make(chan net.Conn)
	flagChan := make(chan bool)

	// 远程拨号，建桥
	go buildTunnelConnection(cfg, index, connChan, flagChan)
	// 本地连接拨号，并建立双向通道
	go buildLocalConnection(cfg.LocalAddr[index], connChan, flagChan)
	// 初始化连接
	for i := 0; i < cfg.TunnelCount; i++ {
		flagChan <- true
	}
	log.Printf("Initilization tunnel [%d] [%d]", cfg.LocalAddr[index].Port2, cfg.TunnelCount)
}

func buildTunnelConnection(cfg config.ClientConfig, index int, connCh chan net.Conn, flagCh chan bool) {
	server := cfg.ServerAddr
	local := cfg.LocalAddr[index]

	for {
		select {
		case <-flagCh:
			// 新建协程，向桥端建立连接
			go func(ch chan net.Conn) {
				conn := dial(server, maxRetryTimes)
				if conn == nil {
					runtime.Goexit()
				}

				request := Protocol{
					Result:  protocolResultSuccess,
					Version: Version,
					Port:    local.Port2,
					ID:      clientID,
					Key:     cfg.Key,
				}

				if !sendProtocol(conn, request) {
					log.Fatalln("Fail to start. exit")
				}
				var response Protocol
			loop:
				for {
					// 此处会阻塞，以等待访问者连接
					response = receiveProtocol(conn)

					// 处理连接结果
					switch response.Result {
					case protocolResultHeartBeat:
						// 不做任何处理，继续监听
						//log.Println("heartbeat...", response.Port)
					case protocolResultSuccess:
						log.Printf("New connection [%d] [%s]\n", local.Port2, local.String())
						ch <- conn

						// 跳出循环
						break loop
					case protocolResultVersionMismatch:
						// 版本不匹配，退出客户端
						// 鉴权失败，退出客户端
						log.Fatalln("Version mismatch. exit")
					case protocolResultFailToAuth:
						// 鉴权失败，退出客户端
						log.Fatalln("Fail to auth. exit")
					case protocolResultIllegalAccessPort:
						// 访问端口不合法
						log.Fatalln("Illegal Access Port. exit")
					case protocolResultPortIsOccupied:
						// 访问端口被占用
						log.Fatalf("Port[%d] is occupied\n", response.Port)
					case protocolResultFail:
						log.Fatalln("Fail to start. exit")
					case protocolResultFailToReceive:
						// 一般是超时导致，不打印日志
						closeConn(conn)
						flagCh <- true

						// 跳出循环
						break loop
					default:
						// 连接中断，重新连接
						log.Printf("Tunnel connection interrupted, try to redial. [result=%d] [%s]\n", response.Result, local.String())
						closeConn(conn)
						flagCh <- true

						// 跳出循环
						break loop
					}
				}
			}(connCh)
		}
	}
}

// 本地服务连接拨号，并建立双向通道
func buildLocalConnection(local config.NetAddress, connCh chan net.Conn, flagCh chan bool) {
	for {
		select {
		case cn := <-connCh:
			// 建立本地连接访问
			go func(conn net.Conn) {
				// 本地连接，不需要重新拨号
				if localConn := dial(local, 0); localConn != nil {
					// 通知创建新桥
					flagCh <- true
					forward(localConn, conn)
				} else {
					// 放弃连接，重新建桥
					closeConn(conn)
					flagCh <- true
				}
			}(cn)
		}
	}
}

// 入口
func Client(cfg config.ClientConfig) {
	log.Println("Load config", cfg)

	// 遍历所有端口
	for index := range cfg.LocalAddr {
		go handleClientConnection(cfg, index)
	}

	select {}
}
