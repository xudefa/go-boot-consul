// Package consul 提供 Consul 配置中心实现
package consul

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"

	"github.com/xudefa/go-boot/config"
)

type ConsulConfigCenter struct {
	client *api.Client
	config *config.ConfigCenterConfig
}

func NewConsulConfigCenter(cfg *config.ConfigCenterConfig) (*ConsulConfigCenter, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, fmt.Errorf("consul endpoints required")
	}

	client, err := api.NewClient(&api.Config{
		Address: cfg.Endpoints[0],
	})
	if err != nil {
		return nil, fmt.Errorf("create consul client failed: %w", err)
	}

	return &ConsulConfigCenter{
		client: client,
		config: cfg,
	}, nil
}

// Load 加载所有配置数据
func (c *ConsulConfigCenter) Load() (config.ConfigData, error) {
	prefix := c.config.Prefix + "/"

	kv := c.client.KV()
	pairs, _, err := kv.List(prefix, &api.QueryOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	result := make(config.ConfigData)
	for _, pair := range pairs {
		relativeKey := strings.TrimPrefix(pair.Key, prefix)
		if relativeKey != "" {
			var value any
			if err := json.Unmarshal(pair.Value, &value); err == nil {
				result[relativeKey] = value
			} else {
				result[relativeKey] = string(pair.Value)
			}
		}
	}

	return result, nil
}

// Watch 监控配置变更
func (c *ConsulConfigCenter) Watch(key string, callback func(config.ConfigData)) error {
	fullKey := c.config.Prefix + "/" + key

	kv := c.client.KV()

	go func() {
		lastIndex := uint64(0)
		for {
			pair, meta, err := kv.Get(fullKey, &api.QueryOptions{
				WaitIndex: lastIndex,
				WaitTime:  30 * time.Second,
			})

			if err != nil {
				continue
			}

			if meta.LastIndex > lastIndex {
				lastIndex = meta.LastIndex
				if pair != nil {
					var data config.ConfigData
					if err := json.Unmarshal(pair.Value, &data); err == nil {
						callback(data)
					}
				}
			}
		}
	}()

	return nil
}

// Close 关闭客户端
func (c *ConsulConfigCenter) Close() error {
	return nil
}
