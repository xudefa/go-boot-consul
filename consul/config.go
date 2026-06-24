package consul

import (
	"time"

	"github.com/xudefa/go-boot/core"
)

// Config Consul 注册中心配置。
type Config struct {
	Address    string
	Datacenter string
	Token      string
	TTL        time.Duration
	Container  core.Container
}

// Option Consul 配置的函数式选项。
type Option func(*Config)

// WithAddress 设置 Consul 服务器地址。
func WithAddress(addr string) Option {
	return func(c *Config) { c.Address = addr }
}

// WithDatacenter 设置 Consul 数据中心。
func WithDatacenter(dc string) Option {
	return func(c *Config) { c.Datacenter = dc }
}

// WithToken 设置 Consul ACL Token。
func WithToken(token string) Option {
	return func(c *Config) { c.Token = token }
}

// WithTTL 设置健康检查 TTL，即服务实例的心跳间隔。
func WithTTL(ttl time.Duration) Option {
	return func(c *Config) { c.TTL = ttl }
}

// WithContainer 设置 IoC 容器实例。
func WithContainer(ctn core.Container) Option {
	return func(c *Config) { c.Container = ctn }
}

func defaultConfig() *Config {
	return &Config{
		Address: "localhost:8500",
		TTL:     10 * time.Second,
	}
}
