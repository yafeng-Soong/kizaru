package router

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gopkg.in/yaml.v3"
)

// route config is a yaml file. content as belows.
// routes:
//   - path: /echo
//     method: POST
//     rpc_method: /echo.EchoService/Echo
type routeInfo struct {
	// {service_name}{path} will be registered into http router.
	Path string `yaml:"path"`
	// the specific gRPC method.
	RPCMethod string `yaml:"rpc_method"`
}

type routeConfig struct {
	Routes []routeInfo `yaml:"routes"`
}

func loadRoutes(ctx context.Context) error {
	// 约定：proto 目录结构为
	// proto/
	//   echo/
	//     echo.proto
	//     echo.yml
	//   foo/
	//     foo.proto
	//     foo.yml

	compiler := protocompile.Compiler{
		Resolver: &protocompile.SourceResolver{
			ImportPaths: []string{"proto"},
		},
	}

	entries, err := os.ReadDir("proto")
	if err != nil {
		return fmt.Errorf("read proto dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// treat dir name as service name
		serviceName := entry.Name()
		if err := loadAppRoutes(ctx, &compiler, serviceName); err != nil {
			return err
		}
	}

	return nil
}

// 加载单个 app 目录下的 proto 与路由配置
func loadAppRoutes(ctx context.Context, compiler *protocompile.Compiler, serviceName string) error {
	var (
		config         routeConfig
		routes         = make(map[string]RouteInfo)
		serviceMethods = make(map[string]protoreflect.MethodDescriptor)
	)

	// read api yml file.
	apiFilePath := path.Join("proto", serviceName, serviceName+".yml")
	apiFile, err := os.ReadFile(apiFilePath)
	if err != nil {
		return fmt.Errorf("read api file %s: %w", apiFilePath, err)
	}

	if err = yaml.Unmarshal(apiFile, &config); err != nil {
		return fmt.Errorf("parse api file %s: %w", apiFilePath, err)
	}

	// 检查该目录下是否包含约定命名的 proto 文件
	protoFilePath := path.Join(serviceName, serviceName+".proto")
	compiled, err := compiler.Compile(ctx, protoFilePath)
	if err != nil {
		return fmt.Errorf("compile proto %s: %w", protoFilePath, err)
	}

	fileDesc := compiled.FindFileByPath(protoFilePath)
	if fileDesc == nil {
		return fmt.Errorf("file not found: %s", protoFilePath)
	}

	services := fileDesc.Services()
	for i := range services.Len() {
		svc := services.Get(i)
		methods := svc.Methods()
		for j := range methods.Len() {
			method := methods.Get(j)
			fullMethod := fmt.Sprintf("/%s/%s", svc.FullName(), method.Name())
			serviceMethods[fullMethod] = method
		}
	}

	for _, route := range config.Routes {
		desc, ok := serviceMethods[route.RPCMethod]
		if !ok {
			return fmt.Errorf("rpc method %s not found", route.RPCMethod)
		}

		apiPath := path.Join(serviceName, route.Path)
		routes[apiPath] = RouteInfo{
			Method: route.RPCMethod,
			Desc:   desc,
		}
	}

	registerRouteInfo(routes)
	return nil
}
