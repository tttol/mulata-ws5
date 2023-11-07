package main

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	"github.com/tttol/mulata-ws5/aws"
)

func main() {
	go runHttp()      // :8000
	go runWebSocket() // :3001
	go runTranslate() // :3002
	select {}
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
	ginEngine.StaticFile("/", "./html/index.html")
	ginEngine.StaticFS("/static", http.Dir("./static"))
	if err := ginEngine.Run(":8000"); err != nil {
		slog.Error("Failed to start Gin server:", err)
	}
	http.ListenAndServe(":8000", nil)
}

func runTranslate() {
	mux := http.NewServeMux()
	mux.HandleFunc("/get/translate", func(w http.ResponseWriter, r *http.Request) {
		result, err := aws.GetResult()
		if err == nil {
			slog.Info("Success to get result: ", result)
			w.Write([]byte(result)) // Write the result to the response body
		} else {
			slog.Error("Failed to get result:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	handler := cors.Default().Handler(mux)
	slog.Info("API server starting on :3002...")
	if err := http.ListenAndServe(":3002", handler); err != nil {
		slog.Error("Failed to start API server:", err)
	}
}
