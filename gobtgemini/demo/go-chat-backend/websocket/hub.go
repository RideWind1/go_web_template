package websocket

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Hub WebSocket连接管理中心
type Hub struct {
	// 已注册的客户端连接
	clients map[*Client]bool

	// 用户ID到客户端的映射
	userClients map[uuid.UUID][]*Client

	// 从客户端接收的消息
	broadcast chan []byte

	// 注册请求
	register chan *Client

	// 注销请求
	unregister chan *Client
}

// NewHub 创建新的Hub
func NewHub() *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[uuid.UUID][]*Client),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
	}
}

// Run 运行Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// 注册客户端
			h.clients[client] = true
			h.userClients[client.userID] = append(h.userClients[client.userID], client)
			logrus.WithFields(logrus.Fields{
				"user_id":   client.userID,
				"client_id": client.id,
				"total":     len(h.clients),
			}).Info("WebSocket客户端已连接")

			// 发送欢迎消息
			welcomeMsg := Message{
				Type:      "system",
				Content:   "欢迎使用智能聊天助手！",
				Timestamp: time.Now(),
			}
			if data, err := json.Marshal(welcomeMsg); err == nil {
				client.send <- data
			}

		case client := <-h.unregister:
			// 注销客户端
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// 从用户客户端列表中删除
				userClients := h.userClients[client.userID]
				for i, c := range userClients {
					if c == client {
						h.userClients[client.userID] = append(userClients[:i], userClients[i+1:]...)
						break
					}
				}

				// 如果用户没有其他连接，删除用户条目
				if len(h.userClients[client.userID]) == 0 {
					delete(h.userClients, client.userID)
				}

				logrus.WithFields(logrus.Fields{
					"user_id":   client.userID,
					"client_id": client.id,
					"total":     len(h.clients),
				}).Info("WebSocket客户端已断开")
			}

		case message := <-h.broadcast:
			// 广播消息给所有客户端
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// SendToUser 发送消息给特定用户
func (h *Hub) SendToUser(userID uuid.UUID, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	userClients, exists := h.userClients[userID]
	if !exists || len(userClients) == 0 {
		logrus.WithField("user_id", userID).Warn("用户无WebSocket连接")
		return nil
	}

	// 发送给用户的所有连接
	for _, client := range userClients {
		select {
		case client.send <- data:
		default:
			// 如果发送失败，关闭连接
			close(client.send)
			delete(h.clients, client)
		}
	}

	return nil
}

// BroadcastMessage 广播消息
func (h *Hub) BroadcastMessage(message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.broadcast <- data
	return nil
}

// GetConnectedUsers 获取在线用户数量
func (h *Hub) GetConnectedUsers() int {
	return len(h.userClients)
}

// GetTotalConnections 获取总连接数
func (h *Hub) GetTotalConnections() int {
	return len(h.clients)
}

// Client WebSocket客户端
type Client struct {
	id       uuid.UUID
	userID   uuid.UUID
	username string
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
}

// Message WebSocket消息结构
type Message struct {
	Type      string      `json:"type"`
	Content   string      `json:"content"`
	UserID    uuid.UUID   `json:"user_id,omitempty"`
	Username  string      `json:"username,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 允许跨域连接
		return true
	},
}

// Handler WebSocket处理器
type Handler struct {
	hub *Hub
}

// NewHandler 创建WebSocket处理器
func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// HandleWebSocket 处理WebSocket连接
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// 从上下文中获取用户信息
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
		return
	}

	username, _ := c.Get("username")

	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.WithError(err).Error("WebSocket升级失败")
		return
	}

	// 创建客户端
	client := &Client{
		id:       uuid.New(),
		userID:   userID.(uuid.UUID),
		username: username.(string),
		hub:      h.hub,
		conn:     conn,
		send:     make(chan []byte, 256),
	}

	// 注册客户端
	client.hub.register <- client

	// 启动客户端的读写协程
	go client.writePump()
	go client.readPump()
}

// readPump 读取消息
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// 设置读取限制
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logrus.WithError(err).Error("意外的WebSocket关闭")
			}
			break
		}

		// 设置消息信息
		msg.UserID = c.userID
		msg.Username = c.username
		msg.Timestamp = time.Now()

		// 处理不同类型的消息
		switch msg.Type {
		case "ping":
			// 回复pong
			pongMsg := Message{
				Type:      "pong",
				Content:   "pong",
				Timestamp: time.Now(),
			}
			if data, err := json.Marshal(pongMsg); err == nil {
				c.send <- data
			}

		case "chat":
			// 这里可以处理聊天消息，但建议通过HTTP API发送
			logrus.WithFields(logrus.Fields{
				"user_id": c.userID,
				"content": msg.Content,
			}).Info("收到聊天消息")

		default:
			logrus.WithField("type", msg.Type).Warn("未知消息类型")
		}
	}
}

// writePump 写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
