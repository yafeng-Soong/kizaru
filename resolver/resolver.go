package resolver

import (
	"flag"
	"log"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	gresolver "google.golang.org/grpc/resolver"
)

var (
	registryAddr = flag.String("registry", "localhost:2379", "")
	etcdClient   *clientv3.Client
	etcdResolver gresolver.Builder
)

func init() {
	flag.Parse()
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: []string{*registryAddr},
	})
	if err != nil {
		log.Fatalf("failed to connect to etcd: %v", err)
	}

	etcdResolver, err = resolver.NewBuilder(etcdClient)
	if err != nil {
		log.Fatalf("failed to create resolver for etcd: %v", err)
	}

	gresolver.Register(etcdResolver)
}

// Scheme returns resolver's scheme.
func Scheme() string {
	return etcdResolver.Scheme()
}

// Clear close etcd client.
func Clear() {
	if etcdClient != nil {
		etcdClient.Close()
	}

	log.Println("[resolver] closed")
}
