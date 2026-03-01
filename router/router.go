package router

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func init() {
	ctx := context.Background()
	if err := loadRoutes(ctx); err != nil {
		log.Fatal(err)
	}
}

var routeMap = make(map[string]RouteInfo)

type RouteInfo struct {
	Method string
	Desc   protoreflect.MethodDescriptor
}

// GetRouteInfo gets route info by http url path.
func GetRouteInfo(urlPath string) (RouteInfo, error) {
	info, ok := routeMap[urlPath]
	if !ok {
		return info, fmt.Errorf("Method not found: %s", urlPath)
	}

	return info, nil
}

func registerRouteInfo(infos map[string]RouteInfo) {
	for path, info := range infos {
		log.Printf("registering route %s -> %s\n", path, info.Method)
		routeMap[path] = info
	}
}
