package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	pb "grpc-echo/proto/echo"
	"grpc-echo/registry"
)

var (
	serviceName = "echo"
	listenAddr  string
)

func init() {
	if localIP := getLocalAddr(); localIP == "" {
		log.Fatal("failed to get local IP address")
	} else {
		listenAddr = fmt.Sprintf("%s:0", localIP)
	}
}

func main() {
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen, error: %v", err)
	}

	ctx := context.Background()
	addr := lis.Addr().String()
	registry.Register(ctx, serviceName, addr)
	defer registry.Deregister(ctx)

	grpcServer := grpc.NewServer()
	pb.RegisterEchoServiceServer(grpcServer, &echoServer{})

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("gRPC echo server listening on %s", addr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-stop
	log.Println("shutting down gRPC server...")
	grpcServer.GracefulStop()
	log.Println("server stopped")
}

type echoServer struct {
	pb.UnimplementedEchoServiceServer
}

func (s *echoServer) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	log.Printf("received: %s", req.GetMessage())
	return &pb.EchoResponse{Message: req.GetMessage()}, nil
}

func getLocalAddr() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip := ipnet.IP.To4(); ip != nil {
				return ip.String()
			}
		}
	}

	return ""
}
