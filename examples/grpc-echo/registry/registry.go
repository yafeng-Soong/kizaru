package registry

import (
	"context"
	"flag"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

const (
	dialTimeout = 5 * time.Second
	logPrefix   = "[registry]"
)

var (
	registryAddr = flag.String("registry", "localhost:2379", "")
	serviceID    string
	client       *clientv3.Client
	manager      endpoints.Manager
)

func init() {
	var err error

	flag.Parse()
	client, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{*registryAddr},
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatalf("%s init etcd client error: %s", logPrefix, err.Error())
	}

	manager, err = endpoints.NewManager(client, "echo")
	if err != nil {
		log.Fatalf("%s init etcd endpoints manager error: %s", logPrefix, err.Error())
	}
}

// Register registers the service to the registry.
func Register(ctx context.Context, service, addr string) {
	serviceID = service + "/" + addr
	if err := manager.AddEndpoint(ctx, serviceID, endpoints.Endpoint{Addr: addr}); err != nil {
		log.Fatalf("%s register error: %s", logPrefix, err.Error())
	}
}

// Deregister deregisters the service from the registry.
func Deregister(ctx context.Context) {
	if err := manager.DeleteEndpoint(ctx, serviceID); err != nil {
		log.Fatalf("%s deregister error: %s", logPrefix, err.Error())
	}

	client.Close()
}
