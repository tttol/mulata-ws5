package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ws", aws.handleConnection)

	fmt.Println("WebSocket server started on :3001...")
	err := http.ListenAndServe(":3001", nil)
	if err != nil {
		log.Fatal("ListenAndServe error on port=3001:", err)
	}
}
