# go-boot-consul

[![Go Version](https://img.shields.io/github/go-mod/go-version/xudefa/go-boot-consul)](https://go.dev/) [![License](https://img.shields.io/github/license/xudefa/go-boot-consul)](./LICENSE) [![Build Status](https://img.shields.io/github/actions/workflow/status/xudefa/go-boot-consul/test.yml?branch=master)](https://github.com/xudefa/go-boot-consul/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/xudefa/go-boot-consul.svg)](https://pkg.go.dev/github.com/xudefa/go-boot-consul) [![Go Report Card](https://goreportcard.com/badge/github.com/xudefa/go-boot-consul)](https://goreportcard.com/report/github.com/xudefa/go-boot-consul)

基于 [go-boot](https://github.com/xudefa/go-boot) 的 Consul 注册中心与配置中心集成模块。将 Consul 无缝集成到 go-boot 的 IoC 容器和自动配置体系中，提供服务注册、服务发现、健康检查和配置管理能力。

> 设计理念：遵循 go-boot 的开发规范，将 Consul 作为 `center.Registry` 和 `config.ConfigCenter` 接口的实现，通过自动配置实现零代码启动服务注册与配置管理。

## 整体架构

```
┌───────────────────────────────────────────────────────────────────────┐
│                    go-boot ApplicationContext                         │
│  ┌───────────┐ ┌──────────────┐ ┌───────────┐ ┌───────────┐           │
│  │ Container │ │  Environment │ │ Lifecycle │ │ EventBus  │           │
│  └───────────┘ └──────────────┘ └───────────┘ └───────────┘           │
│                       ┌─────────────────────┐                         │
│                       │ AutoConfig Registry │                         │
│                       └─────────────────────┘                         │
└───────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │    go-boot-consul Starter     │
                    │  ┌─────────────────────────┐  │
                    │  │ ConsulRegistry Bean     │  │
                    │  │ (center.Registry)       │  │
                    │  │ ConsulConfigCenter Bean │  │
                    │  │ (config.ConfigCenter)   │  │
                    │  │ Health Check & Watch    │  │
                    │  └─────────────────────────┘  │
                    └───────────────────────────────┘
```

## 目录

- [快速开始](#快速开始)
- [功能特性](#功能特性)
- [服务注册与发现](#服务注册与发现)
- [配置中心](#配置中心)
- [配置选项](#配置选项)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
- [贡献](#贡献)
- [许可证](#许可证)

## 快速开始

### 安装

```bash
# 安装核心框架
go get github.com/xudefa/go-boot

# 安装 Consul 集成模块
go get github.com/xudefa/go-boot-consul
```

### 最小示例

```go
package main

import (
    "github.com/xudefa/go-boot/boot"
    "github.com/xudefa/go-boot/center"
)

func main() {
    app, err := boot.NewApplication(
        boot.WithAppName("my-service"),
        boot.WithVersion("1.0.0"),
        boot.WithProperty("consul.enabled", "true"),
        boot.WithProperty("consul.address", "127.0.0.1:8500"),
    )
    if err != nil {
        panic(err)
    }
    defer app.Stop()

    // 获取注册中心（自动注入）
    registry := app.Container().Get("consulRegistry").(center.Registry)

    // 注册服务
    registry.Register(context.Background(), center.InstanceInfo{
        ServiceName: "my-service",
        ID:          "instance-001",
        Host:        "127.0.0.1",
        Port:        8080,
    })

    // 发现服务
    instances, _ := registry.Discover(context.Background(), "my-service")
    for _, inst := range instances {
        fmt.Printf("发现实例: %s:%d\n", inst.Host, inst.Port)
    }

    app.Start()
    app.WaitForSignal()
}
```

## 功能特性

| 特性 | 说明 |
|------|------|
| 注册中心 | 实现 go-boot `center.Registry` 接口 |
| 配置中心 | 实现 go-boot `config.ConfigCenter` 接口 |
| 自动配置 | 通过 `consul.enabled=true` 自动启用 |
| 健康检查 | 支持 TTL 健康检查和 Health API 查询 |
| 阻塞查询 | 使用 Consul Blocking Query 实现长轮询 Watch |
| 数据中心 | 支持多数据中心隔离 |
| 函数式选项 | 灵活的连接配置（Address、Datacenter、Token 等） |

## 服务注册与发现

### 创建注册中心

```go
registry, err := consul.NewConsulRegistry(
    consul.WithAddress("127.0.0.1:8500"),
    consul.WithDatacenter("dc1"),
    consul.WithTTL(30*time.Second),
)
```

### 注册服务

```go
registry.Register(ctx, center.InstanceInfo{
    ServiceName: "user-service",
    ID:          "instance-001",
    Host:        "127.0.0.1",
    Port:        8080,
    Weight:      10,
    Healthy:     true,
    Metadata:    map[string]string{"version": "1.0.0"},
})
```

### 发现服务

```go
instances, err := registry.Discover(ctx, "user-service")
```

### 监听服务变化

```go
ch, err := registry.Watch(ctx, "user-service")
for instances := range ch {
    fmt.Printf("当前在线实例: %d\n", len(instances))
}
```

## 配置中心

### 启用配置中心

```yaml
# application.yml
consul:
  enabled: true
  address: "127.0.0.1:8500"
  config-center:
    enabled: true
    prefix: "config/my-app"
```

### 加载配置

```go
configCenter := app.Container().Get("configCenter").(config.ConfigCenter)
data, err := configCenter.Load()
```

### 监听配置变更

```go
configCenter.Watch("app-config", func(data config.ConfigData) {
    fmt.Printf("配置已更新: %v\n", data)
})
```

## 配置选项

通过 `boot.WithProperty()` 或配置文件设置：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `consul.enabled` | `false` | 是否启用 Consul 注册中心 |
| `consul.address` | `localhost:8500` | Consul 服务端地址 |
| `consul.datacenter` | `` | 数据中心（空则使用服务端默认） |
| `consul.token` | `` | ACL Token |
| `consul.ttl` | `30s` | 健康检查 TTL |
| `consul.config-center.enabled` | `false` | 是否启用配置中心 |
| `consul.config-center.prefix` | `config` | 配置 KV 前缀 |

### 示例配置

```yaml
# application.yml
consul:
  enabled: true
  address: "127.0.0.1:8500"
  datacenter: "dc1"
  ttl: "30s"
  config-center:
    enabled: true
    prefix: "config/my-app"
```

## 项目结构

```
go-boot-consul/
├── consul.go              # ConsulRegistry 实现 center.Registry
├── consul_config.go       # 配置选项（Config, Option）
├── config_center.go       # ConsulConfigCenter 实现 config.ConfigCenter
├── autoconfig.go          # 自动配置注册
├── consul_test.go         # 单元测试
├── README.md
├── LICENSE
└── go.mod
```

## 开发指南

### 构建

```bash
go build ./...
```

### 测试

```bash
go test ./...
go test -cover ./...       # 带覆盖率
go test -race ./...        # 数据竞争检测
```

### 代码规范

```bash
go fmt ./...
golangci-lint run
```

## 贡献

欢迎提交 Issue 和 Pull Request！详细贡献指南请参阅 [CONTRIBUTING.md](./CONTRIBUTING.md)。

## 许可证

本项目采用 MIT 许可证 — 详情请参阅 [LICENSE](./LICENSE) 文件。