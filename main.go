package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"tiny-gateway/rpc"
)

func main() {
	rpc.RegisterHanler()
	defer rpc.Clear()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Println("gRPC gateway is running on port 8080")
		log.Println("Example request: curl -X POST http://localhost:8080/echo/echo -d '{\"message\": \"Hello\"}'")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-signalChan
}
