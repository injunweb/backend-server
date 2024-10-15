package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	notificationWS "github.com/injunweb/backend-server/pkg/websocket"
)

var notificationUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func HandleNotificationWs(c *gin.Context) {
	userId, _ := c.Get("user_id")

	conn, err := notificationUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not open websocket connection"})
		return
	}

	conn.SetReadLimit(512)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	notificationWS.AddNotificationClient(userId.(uint), conn)

	defer notificationWS.RemoveNotificationClient(userId.(uint))

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
