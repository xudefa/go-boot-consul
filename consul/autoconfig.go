// Package consul 提供 Consul 注册中心的自动配置。
//
// 当 consul.enabled=true 时自动启用，从 Environment 中读取 consul.address、consul.datacenter 等配置项，
// 创建并注册 ConsulRegistry Bean 到 IoC 容器中（Bean ID: consulRegistry），实现 center.Registry 接口。
package consul

import (
	"context"
	"fmt"
	"time"

	consulcore "github.com/xudefa/go-boot-consul"

	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/config"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
)

// init 注册 Consul 自动配置和配置中心工厂
func init() {
	boot.RegisterAutoConfig(&ConsulAutoConfiguration{},
		condition.OnProperty("consul.enabled", "true"),
	)

	boot.RegisterConfigCenterFactory("consul", consulConfigCenterFactory)
}

// consulConfigCenterFactory Consul 配置中心工厂函数
func consulConfigCenterFactory(ctx context.Context, cfg *config.ConfigCenterConfig) (config.ConfigCenter, error) {
	return consulcore.NewConsulConfigCenter(cfg)
}

// ConsulAutoConfiguration Consul 注册中心的自动配置。
// 从环境变量读取配置并创建 ConsulRegistry，注册到 IoC 容器中。
// 启用条件：consul.enabled=true
type ConsulAutoConfiguration struct{}

// Configure 执行自动配置逻辑。
// 从 Environment 中读取 consul.address、consul.datacenter 等配置项。
func (c *ConsulAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	// 注册配置中心（如果启用）
	if env.GetBool("consul.config-center.enabled", false) {
		cfg := &config.ConfigCenterConfig{
			Endpoints: []string{env.GetString("consul.address", "localhost:8500")},
			Namespace: env.GetString("consul.datacenter", "dc1"),
			Timeout:   5 * time.Second,
			Prefix:    env.GetString("consul.config-center.prefix", "config"),
		}
		center, err := consulcore.NewConsulConfigCenter(cfg)
		if err != nil {
			return fmt.Errorf("create consul config center failed: %w", err)
		}
		if err := ctx.Register(constants.ConfigCenterBeanID, core.Bean(center), core.Singleton()); err != nil {
			return err
		}
	}

	reg, err := consulcore.NewConsulRegistry(
		consulcore.WithAddress(env.GetString("consul.address", "localhost:8500")),
		consulcore.WithDatacenter(env.GetString("consul.datacenter", "")),
		consulcore.WithContainer(ctx.Container()),
	)
	if err != nil {
		return err
	}

	if err := ctx.Register("consulRegistry",
		core.Bean(reg),
		core.Singleton(),
	); err != nil {
		return err
	}

	return nil
}
