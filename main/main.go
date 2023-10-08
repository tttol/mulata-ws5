package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tttol/mulata-ws5/aws"
)

func main() {
	http.HandleFunc("/ws", aws.HandleConnection)

	fmt.Println("WebSocket server started on :3001...")
	err := http.ListenAndServe(":3001", nil)
	if err != nil {
		log.Fatal("ListenAndServe error on port=3001:", err)
	}
}
