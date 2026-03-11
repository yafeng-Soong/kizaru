// Package main is the entry point of the kizaru gateway.
package main

import (
	"kizaru/rpc"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const defaultReadHeaderTimeout = 3 * time.Second

func main() {
	rpc.RegisterHanler()
	defer rpc.Clear()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: defaultReadHeaderTimeout,
	}

	go func() {
		log.Println("gRPC gateway is running on port 8080")
		log.Println("Example request: curl -X POST http://localhost:8080/echo/echo -d '{\"message\": \"Hello\"}'")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-signalChan
}
