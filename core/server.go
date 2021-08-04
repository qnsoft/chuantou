package core

import (
	"chuantou/config"
	"log"
	"net"
	"sync"
	"time"
)

// 隧道连接
type TunnelConn struct {
	conn       net.Conn
	createTime time.Time
}

// 隧道上下文
type TunnelContext struct {
	request    Protocol        // 请求信息
	listener   net.Listener    // 服务端监听
	tunnelChan chan TunnelConn // 连接池
	createTime time.Time       // 创建时间
	lastTime   time.Time       // 最后检查时间
}

// 存放连接
func (p *TunnelContext) pushConn(conn net.Conn) {
	p.tunnelChan <- TunnelConn{
		conn:       conn,
		createTime: time.Now(),
	}
}

// 心跳，检测连接活性
// 连接池中有连接，则返回成功
func (p *TunnelContext) hearBeat() bool {
	tunnelCount := len(p.tunnelChan)

	for i := 0; i < tunnelCount; i++ {
		tunnelConn := <-p.tunnelChan
		//log.Println("heartbeat...", tunnelConn.createTime.UnixNano())

		// 检测活性
		if sendProtocol(tunnelConn.conn, p.request.NewResult(protocolResultHeartBeat)) {
			// 将连接重新放回连接池
			p.tunnelChan <- tunnelConn
		} else {
			// 关闭失去活性的连接
			closeConn(tunnelConn.conn)
		}
	}
	// 当连接池中没有可用连接时，心跳失败
	p.lastTime = time.Now()
	return len(p.tunnelChan) > 0
}

// key:   accessPort
// value: TunnelContext
var (
	tunnelContextMap   sync.Map
	tunnelContextMutex sync.Mutex
)

// 处理隧道连接
func handleTunnelConnection(tunnelConn net.Conn, cfg config.ServerConfig, tunnelContextChan chan TunnelContext) {
	// 接收协议消息
	req := receiveProtocol(tunnelConn)

	// 检查请求合法性
	if protocolResult := checkRequest(req, cfg); protocolResult != protocolResultSuccess {
		log.Printf("Illegal request, code = %b, ip = %s\n", protocolResult, tunnelConn.RemoteAddr().String())
		sendProtocol(tunnelConn, req.NewResult(protocolResult))
		closeConn(tunnelConn)
		return
	}

	// 获取隧道连接
	context, exists := tunnelContextMap.Load(req.Port)
	if exists {
		tunnelContext := context.(TunnelContext)
		// 端口的开启者是当前访问者
		if tunnelContext.request.IsSameID(&req) {
			tunnelContext.pushConn(tunnelConn)
		} else {
			// 端口已经被其他客户端占用，返回相应提示
			sendProtocol(tunnelConn, req.NewResult(protocolResultPortIsOccupied))
			closeConn(tunnelConn)
		}
		return
	}

	// 第一次创建才会执行，避免每次都加锁
	registerTunnelContext(req, tunnelContextChan)
	tunnelContextMutex.Lock()
	defer tunnelContextMutex.Unlock()
	context, exists = tunnelContextMap.Load(req.Port)
	if !exists {
		context = TunnelContext{
			request:    req,
			listener:   listen(req.Port, req.ID),
			tunnelChan: make(chan TunnelConn, config.MaxTunnelCount),
			createTime: time.Now(),
			lastTime:   time.Now(),
		}
		tunnelContextMap.Store(req.Port, context)
		tunnelContextChan <- context.(TunnelContext)

		log.Printf("Register port [%d] [%s] [%s]\n", req.Port, tunnelConn.RemoteAddr().String(), req.ID)
	}
	tunnelContext := context.(TunnelContext)
	tunnelContext.pushConn(tunnelConn)
}

func registerTunnelContext(req Protocol, tunnelContextChan chan TunnelContext) TunnelContext {
	tunnelContextMutex.Lock()
	defer tunnelContextMutex.Unlock()

	context, exists := tunnelContextMap.Load(req.Port)
	if !exists {
		context = TunnelContext{
			request:    req,
			listener:   listen(req.Port, req.ID),
			tunnelChan: make(chan TunnelConn, config.MaxTunnelCount),
			createTime: time.Now(),
			lastTime:   time.Now(),
		}
		tunnelContextMap.Store(req.Port, context)
		tunnelContextChan <- context.(TunnelContext)
	}
	return context.(TunnelContext)
}

// 检查请求信息，返回结果
func checkRequest(req Protocol, cfg config.ServerConfig) byte {
	if !req.Success() {
		return req.Result
	}
	// 检查版本号
	if req.Version != Version {
		log.Println("Version mismatch", req.String())
		return protocolResultVersionMismatch
	}
	// 检查权限
	if _, ok := config.CheckKey(cfg.Key, req.Key); !ok {
		log.Println("Unauthorized access", req.String())
		return protocolResultFailToAuth
	}
	// 检查访问端口是否在允许范围内
	if ok := cfg.PortInRange(req.Port); !ok {
		log.Println("Access Port out of range", req.String())
		return protocolResultIllegalAccessPort
	}
	return protocolResultSuccess
}

// 处理访问连接
func handleServerConnection(context TunnelContext) {
	serverListener := context.listener
	if serverListener == nil {
		tunnelContextMap.Delete(context.request.Port)
		return
	}
	for {
		serverConn := accept(serverListener)
		if serverConn == nil {
			// 受理监听失败，可能是监听关闭了，结束连接
			break
		}
		// 取隧道连接
		tunnelConn := <-context.tunnelChan
		if sendProtocol(tunnelConn.conn, context.request) {
			log.Printf("Accept connection [%d] [%s]\n", context.request.Port, serverConn.RemoteAddr().String())
			go forward(tunnelConn.conn, serverConn)
		} else {
			log.Printf("No tunnel available, close server listener. [%d]\n", context.request.Port)
			closeConn(serverConn)
			_ = serverListener.Close()
			tunnelContextMap.Delete(context.request.Port)
			break
		}
	}
}

// 入口
func Server(cfg config.ServerConfig) {
	log.Println("Load config", cfg)

	// 监听隧道端口
	tunnelListener := listen(cfg.Port, "server")
	if tunnelListener == nil {
		log.Fatalln("Fail to listen the tunnel port.")
	}

	tunnelContextChan := make(chan TunnelContext)
	// 处理来自客户端的隧道请求
	go func() {
		for {
			tunnelConn := accept(tunnelListener)
			if tunnelConn != nil {
				go handleTunnelConnection(tunnelConn, cfg, tunnelContextChan)
			}
		}
	}()

	// 处理监听及访问
	go func() {
		for {
			select {
			case context := <-tunnelContextChan:
				go handleServerConnection(context)
			}
		}
	}()

	// 心跳，需要考虑端口过多，心跳时间不够的情况
	go setInterval(func() {
		tunnelContextMap.Range(func(key, value interface{}) bool {
			tunnelContext := value.(TunnelContext)
			if !tunnelContext.hearBeat() {
				_ = tunnelContext.listener.Close()
				tunnelContextMap.Delete(key)
			}
			return true
		})
	}, heartBeatIntervalTime)

	select {}
}
