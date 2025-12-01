package voicerecognition

import (
	"diabetes-agent-backend/config"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// WebSocket连接允许的最大空闲时间
	maxIdleTime = 60 * time.Second

	// 连接池大小
	poolSize = 100

	// 语音识别服务URL
	url = "wss://dashscope.aliyuncs.com/api-ws/v1/inference/"
)

type WSConnection struct {
	conn     *websocket.Conn
	lastUsed time.Time
	taskID   string
}

// WSConnectionPool WebSocket连接池
type WSConnectionPool struct {
	pool   chan *WSConnection
	dialer *websocket.Dialer
	url    string
	header http.Header
}

var wsConnectionPool *WSConnectionPool

func init() {
	header := make(http.Header)
	header.Add("Authorization", fmt.Sprintf("bearer %s", config.Cfg.Model.APIKey))
	header.Add("X-DashScope-DataInspection", "enable")
	wsConnectionPool = newWSConnectionPool(url, header)
}

func newWSConnectionPool(url string, header http.Header) *WSConnectionPool {
	return &WSConnectionPool{
		pool:   make(chan *WSConnection, poolSize),
		dialer: websocket.DefaultDialer,
		url:    url,
		header: header,
	}
}

func (p *WSConnectionPool) Get() (*WSConnection, error) {
	for {
		select {
		case conn := <-p.pool:
			// 检测连接是否超时
			if time.Since(conn.lastUsed) > maxIdleTime {
				conn.conn.Close()
				continue
			}
			return conn, nil
		default:
			// 若连接池为空，创建新连接
			conn, _, err := p.dialer.Dial(p.url, p.header)
			if err != nil {
				return nil, err
			}
			return &WSConnection{
				conn:     conn,
				lastUsed: time.Now(),
			}, nil
		}
	}
}

func (p *WSConnectionPool) Put(conn *WSConnection) {
	conn.lastUsed = time.Now()
	select {
	case p.pool <- conn:
	default:
		// 若连接池已满，关闭连接
		conn.conn.Close()
	}
}

func (p *WSConnectionPool) Close() {
	close(p.pool)
	for conn := range p.pool {
		conn.conn.Close()
	}
}
