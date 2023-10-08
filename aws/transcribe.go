package aws

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // この部分は実際の運用時には安全なオリジンチェックを実装するべきです。
	},
}

func HandleConnection(w http.ResponseWriter, r *http.Request) {
	// HTTP接続をWebSocketにアップグレード
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error during connection upgrade:", err)
		return
	}
	defer ws.Close()

	fmt.Println("Client Connected!")
	for {
		messageType, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("Error during reading message:", err)
			break
		}
		fmt.Println("MessageType:", messageType)
		fmt.Println("Received(raw msg):", msg)
		if messageType == 1 {
			fmt.Println("Received(stringfied):", string(msg))
		}

		err = ws.WriteMessage(messageType, msg)
		if err != nil {
			fmt.Println("Error during writing message:", err)
			break
		}
	}
}
