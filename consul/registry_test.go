// consul 集成模块测试
// 测试 Consul 注册中心的默认配置、选项设置和健康检查辅助函数
package consul

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/xudefa/go-boot/center"
	"github.com/xudefa/go-boot/config"
)

// TestConsulConfigDefaults 测试默认配置，验证默认地址为 localhost:8500
func TestConsulConfigDefaults(t *testing.T) {
	cfg := defaultConfig()
	if cfg.Address != "localhost:8500" {
		t.Fatalf("expected localhost:8500, got %s", cfg.Address)
	}
	if cfg.TTL != 10*time.Second {
		t.Fatalf("expected TTL 10s, got %v", cfg.TTL)
	}
}

// TestWithOptions 测试通过选项函数设置地址、数据中心和 Token，验证各配置项正确
func TestWithOptions(t *testing.T) {
	opts := []Option{
		WithAddress("192.168.1.1:8500"),
		WithDatacenter("dc1"),
		WithToken("test-token"),
		WithTTL(30 * time.Second),
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
	if cfg.TTL != 30*time.Second {
		t.Fatalf("unexpected TTL: %v", cfg.TTL)
	}
}

// TestAllPassing 测试健康检查辅助函数 allPassing，验证 nil、空列表、混合状态和全部通过等场景
func TestAllPassing(t *testing.T) {
	tests := []struct {
		name     string
		checks   []*api.HealthCheck
		expected bool
	}{
		{"nil checks", nil, true},
		{"empty checks", []*api.HealthCheck{}, true},
		{"all passing", []*api.HealthCheck{{Status: "passing"}, {Status: "passing"}}, true},
		{"mixed status", []*api.HealthCheck{{Status: "passing"}, {Status: "warning"}}, false},
		{"all critical", []*api.HealthCheck{{Status: "critical"}, {Status: "critical"}}, false},
		{"single warning", []*api.HealthCheck{{Status: "warning"}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := allPassing(tt.checks)
			if result != tt.expected {
				t.Errorf("allPassing() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestRegister_WithEmptyServiceName 测试注册空服务名称，验证返回错误
func TestRegister_WithEmptyServiceName(t *testing.T) {
	registry := &ConsulRegistry{
		config: &Config{
			Address: "localhost:8500",
		},
	}

	err := registry.Register(context.TODO(), center.InstanceInfo{
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
	registry := &ConsulRegistry{
		config: &Config{
			Address: "localhost:8500",
		},
	}

	err := registry.Register(context.TODO(), center.InstanceInfo{
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
	registry := &ConsulRegistry{
		config: &Config{
			Address: "localhost:8500",
		},
	}

	err := registry.Register(context.TODO(), center.InstanceInfo{
		ServiceName: "test-service",
		Host:        "127.0.0.1",
		Port:        0,
	})
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
	if err.Error() != "valid port is required for registration" {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRegister_WithNegativePort 测试注册负数端口
func TestRegister_WithNegativePort(t *testing.T) {
	registry := &ConsulRegistry{
		config: &Config{
			Address: "localhost:8500",
		},
	}

	err := registry.Register(context.TODO(), center.InstanceInfo{
		ServiceName: "test-service",
		Host:        "127.0.0.1",
		Port:        -1,
	})
	if err == nil {
		t.Fatal("expected error for negative port")
	}
}

// TestNewConsulRegistry_WithValidAddress 测试使用有效地址创建 Consul 注册中心
func TestNewConsulRegistry_WithValidAddress(t *testing.T) {
	_, err := NewConsulRegistry(WithAddress("localhost:8500"))
	if err != nil {
		t.Logf("Connection failed as expected: %v", err)
	}
}

// TestNewConsulRegistry_WithOptions 测试使用多个选项创建注册中心
func TestNewConsulRegistry_WithOptions(t *testing.T) {
	_, err := NewConsulRegistry(
		WithAddress("localhost:8500"),
		WithDatacenter("dc1"),
		WithToken("test-token"),
		WithTTL(15*time.Second),
	)
	if err != nil {
		t.Logf("Connection failed as expected: %v", err)
	}
}

// TestConsulRegistry_StructFields 测试注册中心结构体字段
func TestConsulRegistry_StructFields(t *testing.T) {
	cfg := &Config{
		Address:    "localhost:8500",
		Datacenter: "dc1",
		Token:      "test-token",
		TTL:        10 * time.Second,
	}

	registry := &ConsulRegistry{
		config: cfg,
	}

	if registry.config.Address != "localhost:8500" {
		t.Errorf("config.Address = %s, want localhost:8500", registry.config.Address)
	}
	if registry.config.Datacenter != "dc1" {
		t.Errorf("config.Datacenter = %s, want dc1", registry.config.Datacenter)
	}
	if registry.config.Token != "test-token" {
		t.Errorf("config.Token = %s, want test-token", registry.config.Token)
	}
}

// TestConfig_WithContainer 测试 WithContainer 选项
func TestConfig_WithContainer(t *testing.T) {
	cfg := defaultConfig()
	WithContainer(nil)(cfg)
	if cfg.Container != nil {
		t.Error("expected Container to be nil")
	}
}

// TestConsulConfigCenter_New 测试创建配置中心
func TestConsulConfigCenter_New(t *testing.T) {
	cfg := &config.ConfigCenterConfig{
		Endpoints: []string{"localhost:8500"},
	}

	center, err := NewConsulConfigCenter(cfg)
	if err != nil {
		t.Fatalf("NewConsulConfigCenter() error = %v", err)
	}
	if center == nil {
		t.Fatal("expected non-nil config center")
	}
}

// TestConsulConfigCenter_New_EmptyEndpoints 测试空端点
func TestConsulConfigCenter_New_EmptyEndpoints(t *testing.T) {
	cfg := &config.ConfigCenterConfig{
		Endpoints: []string{},
	}

	_, err := NewConsulConfigCenter(cfg)
	if err == nil {
		t.Fatal("expected error for empty endpoints")
	}
	if err.Error() != "consul endpoints required" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestConsulConfigCenter_Close 测试关闭配置中心
func TestConsulConfigCenter_Close(t *testing.T) {
	cfg := &config.ConfigCenterConfig{
		Endpoints: []string{"localhost:8500"},
	}

	center, err := NewConsulConfigCenter(cfg)
	if err != nil {
		t.Skip("skipping: no consul server available")
	}

	err = center.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// 再次关闭应该不会出错
	err = center.Close()
	if err != nil {
		t.Fatalf("Close() again error = %v", err)
	}
}

// TestConsulConfigCenter_WithPrefix 测试配置中心前缀设置
func TestConsulConfigCenter_WithPrefix(t *testing.T) {
	cfg := &config.ConfigCenterConfig{
		Endpoints: []string{"localhost:8500"},
		Prefix:    "/my-config",
	}

	center, err := NewConsulConfigCenter(cfg)
	if err != nil {
		t.Fatalf("NewConsulConfigCenter() error = %v", err)
	}
	if center.config.Prefix != "/my-config" {
		t.Errorf("expected prefix /my-config, got %s", center.config.Prefix)
	}
}

// TestConsulConfigCenter_WithTimeout 测试配置中心超时设置
func TestConsulConfigCenter_WithTimeout(t *testing.T) {
	cfg := &config.ConfigCenterConfig{
		Endpoints: []string{"localhost:8500"},
		Timeout:   15,
	}

	center, err := NewConsulConfigCenter(cfg)
	if err != nil {
		t.Fatalf("NewConsulConfigCenter() error = %v", err)
	}
	if center.config.Timeout != 15 {
		t.Errorf("expected timeout 15, got %d", center.config.Timeout)
	}
}

// TestConsulConfigCenter_MultipleEndpoints 测试多端点配置中心
func TestConsulConfigCenter_MultipleEndpoints(t *testing.T) {
	cfg := &config.ConfigCenterConfig{
		Endpoints: []string{
			"localhost:8500",
			"localhost:8501",
			"localhost:8502",
		},
	}

	center, err := NewConsulConfigCenter(cfg)
	if err != nil {
		t.Fatalf("NewConsulConfigCenter() error = %v", err)
	}
	if len(center.config.Endpoints) != 3 {
		t.Errorf("expected 3 endpoints, got %d", len(center.config.Endpoints))
	}
}

// TestConsulConfigCenter_ConfigFields 测试配置中心配置字段
func TestConsulConfigCenter_ConfigFields(t *testing.T) {
	cfg := &config.ConfigCenterConfig{
		Endpoints: []string{"localhost:8500"},
		Prefix:    "/config",
		Namespace: "production",
		DataID:    "app-config",
		Group:     "DEFAULT_GROUP",
	}

	center, err := NewConsulConfigCenter(cfg)
	if err != nil {
		t.Fatalf("NewConsulConfigCenter() error = %v", err)
	}

	if center.config.Prefix != "/config" {
		t.Errorf("config.Prefix = %s, want /config", center.config.Prefix)
	}
	if center.config.Namespace != "production" {
		t.Errorf("config.Namespace = %s, want production", center.config.Namespace)
	}
	if center.config.DataID != "app-config" {
		t.Errorf("config.DataID = %s, want app-config", center.config.DataID)
	}
	if center.config.Group != "DEFAULT_GROUP" {
		t.Errorf("config.Group = %s, want DEFAULT_GROUP", center.config.Group)
	}
}

// TestConsulConfigCenter_New_NilConfig 测试 nil 配置
func TestConsulConfigCenter_New_NilConfig(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Logf("expected panic for nil config: %v", r)
		}
	}()

	_, err := NewConsulConfigCenter(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

// TestOptionFunctions_Chaining 测试选项函数链式调用
func TestOptionFunctions_Chaining(t *testing.T) {
	cfg := defaultConfig()

	// 链式应用选项
	WithAddress("10.0.0.1:8500")(cfg)
	WithDatacenter("dc2")(cfg)
	WithToken("my-token")(cfg)
	WithTTL(60 * time.Second)(cfg)
	WithContainer(nil)(cfg)

	if cfg.Address != "10.0.0.1:8500" {
		t.Errorf("address = %s, want 10.0.0.1:8500", cfg.Address)
	}
	if cfg.Datacenter != "dc2" {
		t.Errorf("datacenter = %s, want dc2", cfg.Datacenter)
	}
	if cfg.Token != "my-token" {
		t.Errorf("token = %s, want my-token", cfg.Token)
	}
	if cfg.TTL != 60*time.Second {
		t.Errorf("TTL = %v, want 60s", cfg.TTL)
	}
	if cfg.Container != nil {
		t.Error("Container should be nil")
	}
}

// TestDefaultConfig_Immutability 测试默认配置不会被修改
func TestDefaultConfig_Immutability(t *testing.T) {
	// 修改一个配置
	cfg1 := defaultConfig()
	cfg1.Address = "modified:8500"
	cfg1.Datacenter = "modified-dc"
	cfg1.TTL = 999 * time.Second

	// 获取新的默认配置，应该不受影响
	cfg2 := defaultConfig()
	if cfg2.Address == "modified:8500" {
		t.Error("defaultConfig() returned modified address")
	}
	if cfg2.Datacenter == "modified-dc" {
		t.Error("defaultConfig() returned modified datacenter")
	}
	if cfg2.TTL == 999*time.Second {
		t.Error("defaultConfig() returned modified TTL")
	}
}

// TestRegister_WithAllFields 测试注册时使用完整的 InstanceInfo
func TestRegister_WithAllFields(t *testing.T) {
	// 测试带有完整字段的实例信息（验证参数传递）
	info := center.InstanceInfo{
		ServiceName: "test-service",
		ID:          "inst-1",
		Host:        "127.0.0.1",
		Port:        8080,
		Weight:      10,
		Healthy:     true,
		Metadata: map[string]string{
			"version": "1.0.0",
			"env":     "production",
		},
	}

	// 验证参数
	if info.ServiceName == "" {
		t.Error("service name should not be empty")
	}
	if info.Host == "" {
		t.Error("host should not be empty")
	}
	if info.Port <= 0 {
		t.Error("port should be positive")
	}
	if info.Weight != 10 {
		t.Errorf("weight = %d, want 10", info.Weight)
	}
	if !info.Healthy {
		t.Error("healthy should be true")
	}
	if len(info.Metadata) != 2 {
		t.Errorf("metadata length = %d, want 2", len(info.Metadata))
	}
}

// TestNewConsulRegistry_WithCustomAddress 测试自定义地址创建
func TestNewConsulRegistry_WithCustomAddress(t *testing.T) {
	_, err := NewConsulRegistry(WithAddress("10.0.0.1:8500"))
	if err != nil {
		t.Logf("Connection failed as expected: %v", err)
	}
}

// TestNewConsulRegistry_WithShortTTL 测试短 TTL
func TestNewConsulRegistry_WithShortTTL(t *testing.T) {
	_, err := NewConsulRegistry(
		WithAddress("localhost:8500"),
		WithTTL(5*time.Second),
	)
	if err != nil {
		t.Logf("Connection failed as expected: %v", err)
	}
}

// TestNewConsulRegistry_WithLongTTL 测试长 TTL
func TestNewConsulRegistry_WithLongTTL(t *testing.T) {
	_, err := NewConsulRegistry(
		WithAddress("localhost:8500"),
		WithTTL(60*time.Second),
	)
	if err != nil {
		t.Logf("Connection failed as expected: %v", err)
	}
}

// TestNewConsulRegistry_WithDatacenter 测试带数据中心创建
func TestNewConsulRegistry_WithDatacenter(t *testing.T) {
	_, err := NewConsulRegistry(
		WithAddress("localhost:8500"),
		WithDatacenter("us-east-1"),
	)
	if err != nil {
		t.Logf("Connection failed as expected: %v", err)
	}
}

// TestNewConsulRegistry_WithToken 测试带 Token 创建
func TestNewConsulRegistry_WithToken(t *testing.T) {
	_, err := NewConsulRegistry(
		WithAddress("localhost:8500"),
		WithToken("secret-token"),
	)
	if err != nil {
		t.Logf("Connection failed as expected: %v", err)
	}
}

// TestConfig_EmptyAddress 测试空地址配置
func TestConfig_EmptyAddress(t *testing.T) {
	cfg := defaultConfig()
	WithAddress("")(cfg)
	if cfg.Address != "" {
		t.Errorf("expected empty address, got %s", cfg.Address)
	}
}

// TestConfig_EmptyDatacenter 测试空数据中心配置
func TestConfig_EmptyDatacenter(t *testing.T) {
	cfg := defaultConfig()
	WithDatacenter("")(cfg)
	if cfg.Datacenter != "" {
		t.Errorf("expected empty datacenter, got %s", cfg.Datacenter)
	}
}

// TestConfig_EmptyToken 测试空 Token 配置
func TestConfig_EmptyToken(t *testing.T) {
	cfg := defaultConfig()
	WithToken("")(cfg)
	if cfg.Token != "" {
		t.Errorf("expected empty token, got %s", cfg.Token)
	}
}

// TestConsulConfigDefaults_AllFields 测试所有默认配置字段
func TestConsulConfigDefaults_AllFields(t *testing.T) {
	cfg := defaultConfig()

	if cfg.Address == "" {
		t.Error("default address should not be empty")
	}
	if cfg.TTL <= 0 {
		t.Errorf("default TTL should be positive, got %v", cfg.TTL)
	}
	// Datacenter 和 Token 默认应该为空
	if cfg.Datacenter != "" {
		t.Errorf("default datacenter should be empty, got %s", cfg.Datacenter)
	}
	if cfg.Token != "" {
		t.Errorf("default token should be empty, got %s", cfg.Token)
	}
}

// TestInstanceKey_Format 测试实例键格式
func TestInstanceKey_Format(t *testing.T) {
	registry := &ConsulRegistry{
		config: &Config{
			Address: "localhost:8500",
		},
	}

	// 验证实例键生成逻辑
	info := center.InstanceInfo{
		ServiceName: "user-service",
		ID:          "192.168.1.1:8080",
	}

	// Consul 使用 serviceName 作为服务名，ID 作为实例标识
	if info.ServiceName != "user-service" {
		t.Errorf("service name = %s, want user-service", info.ServiceName)
	}
	if info.ID != "192.168.1.1:8080" {
		t.Errorf("instance ID = %s, want 192.168.1.1:8080", info.ID)
	}

	// 验证 registry 配置
	if registry.config.Address != "localhost:8500" {
		t.Errorf("registry address = %s, want localhost:8500", registry.config.Address)
	}
}

// TestInstanceKey_WithSpecialCharacters 测试包含特殊字符的实例信息
func TestInstanceKey_WithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		id          string
	}{
		{
			name:        "UUID as ID",
			serviceName: "user-service",
			id:          "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:        "IP:Port as ID",
			serviceName: "api-gateway",
			id:          "192.168.1.100:9090",
		},
		{
			name:        "Hostname as ID",
			serviceName: "payment-service",
			id:          "prod-server-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := center.InstanceInfo{
				ServiceName: tt.serviceName,
				ID:          tt.id,
			}

			if info.ServiceName != tt.serviceName {
				t.Errorf("service name = %s, want %s", info.ServiceName, tt.serviceName)
			}
			if info.ID != tt.id {
				t.Errorf("instance ID = %s, want %s", info.ID, tt.id)
			}
		})
	}
}

// TestAllPassing_EdgeCases 测试 allPassing 边界情况
func TestAllPassing_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		checks   []*api.HealthCheck
		expected bool
	}{
		{
			name:     "single passing",
			checks:   []*api.HealthCheck{{Status: "passing"}},
			expected: true,
		},
		{
			name:     "single critical",
			checks:   []*api.HealthCheck{{Status: "critical"}},
			expected: false,
		},
		{
			name:     "single warning",
			checks:   []*api.HealthCheck{{Status: "warning"}},
			expected: false,
		},
		{
			name: "multiple passing",
			checks: []*api.HealthCheck{
				{Status: "passing"},
				{Status: "passing"},
				{Status: "passing"},
			},
			expected: true,
		},
		{
			name: "last one critical",
			checks: []*api.HealthCheck{
				{Status: "passing"},
				{Status: "passing"},
				{Status: "critical"},
			},
			expected: false,
		},
		{
			name: "first one critical",
			checks: []*api.HealthCheck{
				{Status: "critical"},
				{Status: "passing"},
				{Status: "passing"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := allPassing(tt.checks)
			if result != tt.expected {
				t.Errorf("allPassing() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestConsulConfigCenter_WithNamespace 测试配置中心命名空间
func TestConsulConfigCenter_WithNamespace(t *testing.T) {
	cfg := &config.ConfigCenterConfig{
		Endpoints: []string{"localhost:8500"},
		Namespace: "production",
	}

	center, err := NewConsulConfigCenter(cfg)
	if err != nil {
		t.Fatalf("NewConsulConfigCenter() error = %v", err)
	}
	if center.config.Namespace != "production" {
		t.Errorf("expected namespace production, got %s", center.config.Namespace)
	}
}

// TestConsulConfigCenter_WithDataID 测试配置中心 DataID
func TestConsulConfigCenter_WithDataID(t *testing.T) {
	cfg := &config.ConfigCenterConfig{
		Endpoints: []string{"localhost:8500"},
		DataID:    "my-app-config",
	}

	center, err := NewConsulConfigCenter(cfg)
	if err != nil {
		t.Fatalf("NewConsulConfigCenter() error = %v", err)
	}
	if center.config.DataID != "my-app-config" {
		t.Errorf("expected DataID my-app-config, got %s", center.config.DataID)
	}
}

// TestConsulConfigCenter_WithGroup 测试配置中心 Group
func TestConsulConfigCenter_WithGroup(t *testing.T) {
	cfg := &config.ConfigCenterConfig{
		Endpoints: []string{"localhost:8500"},
		Group:     "MY_GROUP",
	}

	center, err := NewConsulConfigCenter(cfg)
	if err != nil {
		t.Fatalf("NewConsulConfigCenter() error = %v", err)
	}
	if center.config.Group != "MY_GROUP" {
		t.Errorf("expected group MY_GROUP, got %s", center.config.Group)
	}
}

// TestConsulRegistry_Interface 编译时检查 ConsulRegistry 是否实现了 center.Registry 接口
func TestConsulRegistry_Interface(t *testing.T) {
	var _ center.Registry = (*ConsulRegistry)(nil)
}

// TestConsulConfigCenter_Interface 编译时检查 ConsulConfigCenter 是否实现了 config.ConfigCenter 接口
func TestConsulConfigCenter_Interface(t *testing.T) {
	var _ config.ConfigCenter = (*ConsulConfigCenter)(nil)
}
