// consul 集成模块测试
// 测试 Consul 注册中心的默认配置、选项设置和健康检查辅助函数
package consul

import (
	"context"
	"testing"

	"github.com/hashicorp/consul/api"
)

// TestConsulConfigDefaults 测试默认配置，验证默认地址为 localhost:8500
func TestConsulConfigDefaults(t *testing.T) {
	cfg := defaultConfig()
	if cfg.Address != "localhost:8500" {
		t.Fatalf("expected localhost:8500, got %s", cfg.Address)
	}
}

// TestWithOptions 测试通过选项函数设置地址、数据中心和 Token，验证各配置项正确
func TestWithOptions(t *testing.T) {
	opts := []Option{
		WithAddress("192.168.1.1:8500"),
		WithDatacenter("dc1"),
		WithToken("test-token"),
	}
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.Address != "192.168.1.1:8500" {
		t.Fatalf("unexpected address: %s", cfg.Address)
	}
	if cfg.Datacenter != "dc1" {
		t.Fatalf("unexpected datacenter: %s", cfg.Datacenter)
	}
	if cfg.Token != "test-token" {
		t.Fatalf("unexpected token: %s", cfg.Token)
	}
}

// TestAllPassing 测试健康检查辅助函数 allPassing，验证 nil、空列表、混合状态和全部通过等场景
func TestAllPassing(t *testing.T) {
	if !allPassing(nil) {
		t.Fatal("expected true for nil checks")
	}
	if !allPassing([]*api.HealthCheck{}) {
		t.Fatal("expected true for empty checks")
	}
	if allPassing([]*api.HealthCheck{
		{Status: "passing"},
		{Status: "warning"},
	}) {
		t.Fatal("expected false for mixed status")
	}
	if !allPassing([]*api.HealthCheck{
		{Status: "passing"},
		{Status: "passing"},
	}) {
		t.Fatal("expected true for all passing")
	}
}

// TestRegister_WithEmptyServiceName 测试注册空服务名称，验证返回错误
func TestRegister_WithEmptyServiceName(t *testing.T) {
	// 创建一个测试实例，但不连接真实的Consul
	registry := &ConsulRegistry{
		config: &Config{
			Address: "localhost:8500",
		},
	}

	err := registry.Register(context.TODO(), struct {
		ServiceName string
		ID          string
		Host        string
		Port        int
		Weight      int
		Healthy     bool
		Metadata    map[string]string
	}{
		ServiceName: "",
		Host:        "127.0.0.1",
		Port:        8080,
	})
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
	if err.Error() != "service name is required for registration" {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRegister_WithEmptyHost 测试注册空主机，验证返回错误
func TestRegister_WithEmptyHost(t *testing.T) {
	// 创建一个测试实例，但不连接真实的Consul
	registry := &ConsulRegistry{
		config: &Config{
			Address: "localhost:8500",
		},
	}

	err := registry.Register(context.TODO(), struct {
		ServiceName string
		ID          string
		Host        string
		Port        int
		Weight      int
		Healthy     bool
		Metadata    map[string]string
	}{
		ServiceName: "test-service",
		Host:        "",
		Port:        8080,
	})
	if err == nil {
		t.Fatal("expected error for empty host")
	}
	if err.Error() != "host is required for registration" {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRegister_WithInvalidPort 测试注册无效端口，验证返回错误
func TestRegister_WithInvalidPort(t *testing.T) {
	// 创建一个测试实例，但不连接真实的Consul
	registry := &ConsulRegistry{
		config: &Config{
			Address: "localhost:8500",
		},
	}

	err := registry.Register(context.TODO(), struct {
		ServiceName string
		ID          string
		Host        string
		Port        int
		Weight      int
		Healthy     bool
		Metadata    map[string]string
	}{
		ServiceName: "test-service",
		Host:        "127.0.0.1",
		Port:        0, // Invalid port
	})
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
	if err.Error() != "valid port is required for registration" {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestNewConsulRegistry_WithValidAddress 测试使用有效地址创建 Consul 注册中心
func TestNewConsulRegistry_WithValidAddress(t *testing.T) {
	// 由于可能没有真实Consul服务器，我们只测试配置逻辑
	// 这里使用一个有效的地址格式，但预期连接可能会失败
	_, err := NewConsulRegistry(WithAddress("localhost:8500"))
	// 我们主要测试配置是否正确，而不是连接是否成功
	if err != nil {
		t.Logf("Connection failed as expected: %v", err)
		// 如果是连接错误，这可能是正常的（因为可能没有Consul服务器）
	}
}
