// Package consul 基于 Consul 提供注册中心实现。
//
// 该包将 Consul 与 go-boot 服务发现和注册中心接口集成，
// 支持服务注册、发现和阻塞查询（Watch）。
//
// 定义：
//
//   - ConsulRegistry: 注册中心实现了 center.Registry 接口
//   - Config: Consul 配置
//   - Option: 配置选项函数
//
// 快速开始:
//
//	// 创建 Consul 注册中心
//	registry, err := consul.NewConsulRegistry(
//	    consul.WithAddress("127.0.0.1:8500"),
//	)
//
//	// 注册服务
//	registry.Register(ctx, center.InstanceInfo{
//	    ServiceName: "my-service",
//	    Host:        "127.0.0.1",
//	    Port:        8080,
//	})
package consul

import (
	"context"
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/xudefa/go-boot/center"
)

// ConsulRegistry 基于 Consul 实现的注册中心。
// 使用 Consul HTTP API 实现服务注册、发现和阻塞查询（Watch）。
type ConsulRegistry struct {
	client *api.Client
	config *Config
}

// NewConsulRegistry 创建 Consul 注册中心实例。
// 支持通过 Option 配置地址、数据中心、Token 等参数。
func NewConsulRegistry(opts ...Option) (*ConsulRegistry, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	consulCfg := api.DefaultConfig()
	consulCfg.Address = cfg.Address
	if cfg.Datacenter != "" {
		consulCfg.Datacenter = cfg.Datacenter
	}
	if cfg.Token != "" {
		consulCfg.Token = cfg.Token
	}

	client, err := api.NewClient(consulCfg)
	if err != nil {
		return nil, fmt.Errorf("consul: create client failed: %w", err)
	}

	return &ConsulRegistry{client: client, config: cfg}, nil
}

// Register 向 Consul 注册一个服务实例。
// 如果配置了 TTL，同时注册健康检查。
// 权重信息存储在 Meta["weight"] 中。
func (r *ConsulRegistry) Register(_ context.Context, info center.InstanceInfo) error {
	if info.ServiceName == "" {
		return fmt.Errorf("service name is required for registration")
	}
	if info.Host == "" {
		return fmt.Errorf("host is required for registration")
	}
	if info.Port <= 0 {
		return fmt.Errorf("valid port is required for registration")
	}

	reg := &api.AgentServiceRegistration{
		ID:      info.ID,
		Name:    info.ServiceName,
		Address: info.Host,
		Port:    info.Port,
		Meta:    info.Metadata,
	}
	if info.Weight > 0 {
		if reg.Meta == nil {
			reg.Meta = make(map[string]string)
		}
		reg.Meta["weight"] = fmt.Sprintf("%d", info.Weight)
	}

	if r.config.TTL > 0 {
		reg.Check = &api.AgentServiceCheck{
			TTL:                            r.config.TTL.String(),
			DeregisterCriticalServiceAfter: "1m",
			Status:                         "passing",
		}
	}

	if err := r.client.Agent().ServiceRegister(reg); err != nil {
		return fmt.Errorf("failed to register service to Consul: %w", err)
	}
	return nil
}

// Deregister 从 Consul 注销指定服务实例。
func (r *ConsulRegistry) Deregister(_ context.Context, info center.InstanceInfo) error {
	return r.client.Agent().ServiceDeregister(info.ID)
}

// Discover 发现指定服务的健康实例列表。
// 通过 Consul Health API 查询，仅返回通过健康检查的实例。
func (r *ConsulRegistry) Discover(_ context.Context, serviceName string) ([]center.InstanceInfo, error) {
	entries, _, err := r.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}

	instances := make([]center.InstanceInfo, 0, len(entries))
	for _, entry := range entries {
		svc := entry.Service
		info := center.InstanceInfo{
			ServiceName: svc.Service,
			ID:          svc.ID,
			Host:        svc.Address,
			Port:        svc.Port,
			Metadata:    svc.Meta,
			Healthy:     len(entry.Checks) == 0 || allPassing(entry.Checks),
		}
		if w, ok := svc.Meta["weight"]; ok {
			if _, err := fmt.Sscanf(w, "%d", &info.Weight); err != nil {
				info.Weight = 1
			}
		}
		instances = append(instances, info)
	}
	return instances, nil
}

// Watch 监听指定服务的实例变化。
// 使用 Consul 的阻塞查询（Blocking Query）实现长轮询，
// 通过 WaitIndex 增量获取变化。Watch 的返回结果与 Discover 完全一致。
func (r *ConsulRegistry) Watch(ctx context.Context, serviceName string) (<-chan []center.InstanceInfo, error) {
	ch := make(chan []center.InstanceInfo, 16)

	go func() {
		defer close(ch)
		lastIndex := uint64(0)
		for {
			q := &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  r.config.TTL,
			}
			entries, meta, err := r.client.Health().Service(serviceName, "", true, q)
			if err != nil {
				return
			}
			lastIndex = meta.LastIndex

			instances := make([]center.InstanceInfo, 0, len(entries))
			for _, entry := range entries {
				svc := entry.Service
				info := center.InstanceInfo{
					ServiceName: svc.Service,
					ID:          svc.ID,
					Host:        svc.Address,
					Port:        svc.Port,
					Metadata:    svc.Meta,
					Healthy:     len(entry.Checks) == 0 || allPassing(entry.Checks),
				}
				if w, ok := svc.Meta["weight"]; ok {
					if _, err := fmt.Sscanf(w, "%d", &info.Weight); err != nil {
						info.Weight = 1
					}
				}
				instances = append(instances, info)
			}

			select {
			case ch <- instances:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

// allPassing 检查 Consul 健康检查列表是否全部为 passing 状态。
func allPassing(checks []*api.HealthCheck) bool {
	if len(checks) == 0 {
		return true
	}
	for _, check := range checks {
		if check.Status != "passing" {
			return false
		}
	}
	return true
}
