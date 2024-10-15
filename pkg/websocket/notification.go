package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	notificationClients   = make(map[uint]*websocket.Conn)
	notificationMutex     = &sync.RWMutex{}
	notificationPool      = make(map[uint][]string)
	notificationPoolMutex = &sync.RWMutex{}
	pingInterval          = 30 * time.Second
	writeWait             = 10 * time.Second
)

func AddNotificationClient(userId uint, conn *websocket.Conn) {
	notificationMutex.Lock()
	notificationClients[userId] = conn
	notificationMutex.Unlock()

	go pingNotificationClient(userId, conn)

	notificationPoolMutex.Lock()
	if notifications, ok := notificationPool[userId]; ok {
		for _, notification := range notifications {
			conn.WriteMessage(websocket.TextMessage, []byte(notification))
		}
		delete(notificationPool, userId)
	}
	notificationPoolMutex.Unlock()
}

func RemoveNotificationClient(userId uint) {
	notificationMutex.Lock()
	if conn, ok := notificationClients[userId]; ok {
		conn.Close()
		delete(notificationClients, userId)
	}
	notificationMutex.Unlock()
}

func SendNotification(userId uint, message string) {
	notificationMutex.RLock()
	conn, exists := notificationClients[userId]
	notificationMutex.RUnlock()

	if exists {
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		err := conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("Error sending notification to user %d: %v", userId, err)
			RemoveNotificationClient(userId)
			addToNotificationPool(userId, message)
		}
	} else {
		addToNotificationPool(userId, message)
	}
}

func addToNotificationPool(userId uint, message string) {
	notificationPoolMutex.Lock()
	notificationPool[userId] = append(notificationPool[userId], message)
	notificationPoolMutex.Unlock()
}

func pingNotificationClient(userId uint, conn *websocket.Conn) {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		RemoveNotificationClient(userId)
	}()

	for {
		<-ticker.C
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
	}
}
