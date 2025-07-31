package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境需要限制
	},
}

// WebSocketHandler WebSocket连接处理器
func WebSocketHandler(c *gin.Context) {
	// 升级HTTP连接为WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 发送连接成功消息
	err = conn.WriteMessage(websocket.TextMessage, []byte("WebSocket connected"))
	if err != nil {
		return
	}

	// 处理WebSocket消息
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// 回显消息
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			break
		}
	}
} 