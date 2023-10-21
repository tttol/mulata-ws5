package main

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tttol/mulata-ws5/aws"
)

func main() {
	runWebSocket()
	runHttp()
}

func runWebSocket() {
	http.HandleFunc("/ws", aws.HandleConnection)

	// Run WebSocket server in a goroutine
	go func() {
		slog.Info("WebSocket server started on :3001...")
		if err := http.ListenAndServe(":3001", nil); err != nil {
			slog.Error("ListenAndServe error on port 3001:", err)
		}
	}()
}

func runHttp() {
	ginEngine := gin.Default()
	ginEngine.StaticFile("/", "./static/index.html")
	if err := ginEngine.Run(":8000"); err != nil {
		slog.Error("Failed to start Gin server:", err)
	}
	slog.Info("HTTP server started on :8000...")
}
