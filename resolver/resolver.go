// Package resolver provides etcd-based gRPC resolver implementation.
package resolver

import (
	"flag"
	"log"

	"github.com/yafeng-Soong/aokiji/resolver"
	clientv3 "go.etcd.io/etcd/client/v3"
	gresolver "google.golang.org/grpc/resolver"
)

var (
	registryAddr = flag.String("registry", "localhost:2379", "")
	etcdResolver gresolver.Builder
)

func init() {
	flag.Parse()
	var err error
	etcdResolver, err = resolver.NewEtcdBuilder(clientv3.Config{
		Endpoints: []string{*registryAddr},
	})
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
	log.Println("[resolver] closed")
}
