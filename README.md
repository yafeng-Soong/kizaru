# kizaru

一个用于学习的轻量级 **API 网关 Demo**，主要功能是：

- **接收外部 HTTP + JSON 请求**
- **根据路由配置找到对应的 gRPC 方法**
- **完成 JSON ⇄ Protobuf 的编解码**
- **通过 gRPC 调用后端服务并返回 JSON 响应**

目前项目中已经包含一个基于 `echo` 服务的最小可运行示例，方便你理解整体链路。

---

## 特性概览

- **按约定自动加载路由**
  - 从 `proto` 目录加载 `.proto` 文件
  - 从同名 `.yml` 文件加载 HTTP 路由配置
  - 在启动时构建 URL Path → gRPC Method 的映射表

- **动态 Protobuf 处理**
  - 使用 `protocompile` 编译 `.proto`
  - 使用 `dynamicpb` + `protojson` 在 **JSON 与 Protobuf 消息之间转换**

- **服务发现与负载均衡（基于 etcd）**
  - 使用 etcd 作为 gRPC 的服务发现注册中心
  - 自定义 `resolver`，通过 `scheme://service-name` 形式连接
  - 使用 gRPC 的 `round_robin` 负载均衡策略

- **学习友好**
  - 代码尽量保持简单直接，方便你逐步阅读
  - 只实现了最小可用功能，便于你在此基础上扩展

---

## 目录结构（核心部分）

```text
kizaru/
  main.go              # 网关入口，启动 HTTP Server
  rpc/
    rpc.go             # HTTP 处理逻辑，转发到 gRPC
  router/
    loader.go          # 从 proto + yml 加载路由配置
    router.go          # 路由信息的注册与查询
  resolver/
    resolver.go        # etcd 作为服务发现的 resolver
  proto/
    echo/
      echo.proto       # 示例 gRPC 服务定义（需自行准备）
      echo.yml         # 示例 HTTP 路由配置
```

> 说明：仓库中目前只包含 `echo.yml`，`echo.proto` 和对应的 gRPC 服务实现需要你根据自己的实验环境补上。

---

## 路由配置说明

网关通过 `proto/<service>/<service>.yml` 文件加载 HTTP 路由信息，例如当前示例：

```yaml
app_name: echo
routes:
  - path: /echo
    method: POST
    rpc_method: /echo.EchoService/Echo
```

- **`path`**: HTTP 路径（这里会与 `serviceName` 拼接，形成最终访问路径）
- **`method`**: HTTP Method（目前网关内部只接受 `POST`）
- **`rpc_method`**: 完整的 gRPC 方法名，格式为 `/包名.服务名/方法名`

在代码中，会把目录名视为 `serviceName`，因此：

- 目录为 `proto/echo/`
- HTTP 实际访问路径为：`/echo/echo`
- 对应的 gRPC 方法为：`/echo.EchoService/Echo`

路由加载大致流程在 `router/loader.go` 中：

- 扫描 `proto` 目录下所有子目录
- 对每个 `<service>`：
  - 读取 `<service>/<service>.yml` → 加载 routes
  - 编译 `<service>/<service>.proto` → 获取 gRPC Method 描述
  - 将 `serviceName + path` 映射到对应的 gRPC Method

---

## 运行示例

### 1. 准备环境

- **Go** 1.26+
- **etcd**（默认地址：`localhost:2379`）
- 一个已经实现并 **注册到 etcd** 的 gRPC `echo` 服务，满足：
  - proto 文件路径：`proto/echo/echo.proto`
  - gRPC 服务名：`echo.EchoService`
  - 方法：`Echo`
  - 注册到 etcd 时的 service name 为：`echo`

> - examples/grpc-echo里面有一个满足条件的最小echo server实现
> - 如果你已有自己的 gRPC 服务，也可以按相同约定调整目录与配置。

### 2. 启动 etcd

```bash
etcd
```

如果你的 etcd 不在本地 `localhost:2379`，可以通过参数传入：

```bash
go run . --registry <your-etcd-host:port>
```

### 3. 启动你的 gRPC 后端服务

确保：

- 它使用 gRPC 官方 etcd resolver 或等价方式进行服务注册
- service name 与本项目中 `serviceName` 一致（例如 `echo`）

### 4. 启动网关

在项目根目录执行：

```bash
go run .
```

你将看到类似日志：

```text
gRPC gateway is running on port 8080
Example request: curl -X POST http://localhost:8080/echo/echo -d '{"message": "Hello"}'
```

---

## 调用示例

当网关与后端 gRPC 服务都正常启动后，可以通过 curl 访问：

```bash
curl -X POST \
  http://localhost:8080/echo/echo \
  -H 'Content-Type: application/json' \
  -d '{"message": "Hello"}'
```

网关会做的事情：

1. 接收到 HTTP JSON 请求
2. 根据路径 `echo/echo` 在路由表中查找对应的 gRPC 方法
3. 使用动态 Protobuf 将 JSON 转为请求消息
4. 通过 gRPC 调用后端 `echo.EchoService/Echo`
5. 将返回的 Protobuf 响应再编码为 JSON，返回给客户端

---

## 学习建议与扩展方向

你可以在此基础上尝试：

- **增加更多服务与路由**
  - 在 `proto/<service>` 目录下添加新的 `.proto` 与 `.yml`
  - 观察路由加载和 gRPC 调用链路

- **扩展功能**
  - 增加鉴权 / 鉴权中间件
  - 增加请求/响应日志、Tracing
  - 支持更多 HTTP Method 或 Path 参数映射

- **对比现有成熟方案**
  - 如 `grpc-gateway`、`Envoy` / `APISIX` 等
  - 思考本项目与它们在设计上的差异

本仓库主要用于 **理解“HTTP → gRPC”网关的基本实现思路**，不建议直接用于生产环境。如果你在阅读或扩展过程中遇到问题，可以在代码相应目录里加上自己的注释和实验用例，加深理解。

