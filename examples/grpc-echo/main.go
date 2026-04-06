package main

import (
	"context"
	"flag"
	"log"

	"google.golang.org/grpc"

	pb "grpc-echo/proto/echo"

	"github.com/yafeng-Soong/aokiji/registry/etcd"
	"github.com/yafeng-Soong/aokiji/server"
)

var registryAddr = flag.String("registry", "localhost:2379", "etcd registry address")

func init() {
	flag.Parse()
}

func main() {
	grpcServer := grpc.NewServer()
	pb.RegisterEchoServiceServer(grpcServer, &echoServer{})

	etcdRegistry, err := etcd.NewRegistry(
		etcd.WithEndpoints([]string{*registryAddr}),
	)
	if err != nil {
		log.Fatalf("failed to create etcd registry: %v", err)
	}

	svr := server.NewServer(
		server.WithServiceName("echo"),
		server.WithGRPCServer(grpcServer),
		server.WithRegistry(etcdRegistry),
	)

	if err := svr.Start(context.Background()); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

type echoServer struct {
	pb.UnimplementedEchoServiceServer
}

func (s *echoServer) Echo(ctx context.Context, req *pb.EchoRequest) (*pb.EchoResponse, error) {
	log.Printf("received: %s", req.GetMessage())
	return &pb.EchoResponse{Message: req.GetMessage()}, nil
}
