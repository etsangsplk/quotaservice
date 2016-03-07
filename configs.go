// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

// Package implements configs for the quotaservice
package quotaservice

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/maniksurtani/quotaservice/logging"
	"gopkg.in/yaml.v2"
)

type ServiceConfig struct {
	GlobalDefaultBucket *BucketConfig               `yaml:"global_default_bucket,flow"`
	ListenerBufferSize  int                         `yaml:"listener_buffer_size,flow"`
	Namespaces          map[string]*NamespaceConfig `yaml:",flow"`
}

func (s *ServiceConfig) String() string {
	return fmt.Sprintf("ServiceConfig{default: %v, namespaces: %v}",
		s.GlobalDefaultBucket, s.Namespaces)
}

func (s *ServiceConfig) AddNamespace(namespace string, n *NamespaceConfig) *ServiceConfig {
	s.Namespaces[namespace] = n
	return s
}

func (s *ServiceConfig) applyDefaults() *ServiceConfig {
	if s.GlobalDefaultBucket != nil {
		s.GlobalDefaultBucket.applyDefaults()
	}

	for name, ns := range s.Namespaces {
		if ns.DefaultBucket != nil && ns.DynamicBucketTemplate != nil {
			panic(fmt.Sprintf("Namespace %v is not allowed to have a default bucket as well as allow dynamic buckets.", name))
		}

		// Ensure the namespace's bucket map exists.
		if ns.Buckets == nil {
			ns.Buckets = make(map[string]*BucketConfig)
		}

		if ns.DefaultBucket != nil {
			ns.DefaultBucket.applyDefaults()
		}

		if ns.DynamicBucketTemplate != nil {
			ns.DynamicBucketTemplate.applyDefaults()
		}

		for _, b := range ns.Buckets {
			b.applyDefaults()
		}
	}

	return s
}

type NamespaceConfig struct {
	DefaultBucket         *BucketConfig            `yaml:"default_bucket,flow"`
	DynamicBucketTemplate *BucketConfig            `yaml:"dynamic_bucket_template,flow"`
	MaxDynamicBuckets     int                      `yaml:"max_dynamic_buckets"`
	Buckets               map[string]*BucketConfig `yaml:",flow"`
}

func (n *NamespaceConfig) AddBucket(name string, b *BucketConfig) *NamespaceConfig {
	n.Buckets[name] = b
	return n
}

type BucketConfig struct {
	Size                int64
	FillRate            int64 `yaml:"fill_rate"`
	WaitTimeoutMillis   int64 `yaml:"wait_timeout_millis"`
	MaxIdleMillis       int64 `yaml:"max_idle_millis"`
	MaxDebtMillis       int64 `yaml:"max_debt_millis"`
	MaxTokensPerRequest int64 `yaml:"max_tokens_per_request"`
}

func (b *BucketConfig) String() string {
	return fmt.Sprint(*b)
}

func (b *BucketConfig) applyDefaults() *BucketConfig {
	if b.Size == 0 {
		b.Size = 100
	}

	if b.FillRate == 0 {
		b.FillRate = 50
	}

	if b.WaitTimeoutMillis == 0 {
		b.WaitTimeoutMillis = 1000
	}

	if b.MaxIdleMillis == 0 {
		b.MaxIdleMillis = -1
	}

	if b.MaxDebtMillis == 0 {
		b.MaxDebtMillis = 10000
	}

	if b.MaxTokensPerRequest == 0 {
		b.MaxTokensPerRequest = b.FillRate
	}

	return b
}

func ReadConfigFromFile(filename string) *ServiceConfig {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("Unable to open file %v. Error: %v", filename, err))
	}

	return readConfigFromBytes(bytes)
}

func ReadConfig(yamlStream io.Reader) *ServiceConfig {
	bytes, err := ioutil.ReadAll(yamlStream)
	if err != nil {
		panic(fmt.Sprintf("Unable to open reader. Error: %v", err))
	}

	return readConfigFromBytes(bytes)
}

func readConfigFromBytes(bytes []byte) *ServiceConfig {
	logging.Print(string(bytes))
	cfg := NewDefaultServiceConfig()
	cfg.GlobalDefaultBucket = nil
	yaml.Unmarshal(bytes, cfg)

	return cfg.applyDefaults()
}

func NewDefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		NewDefaultBucketConfig(),
		10000,
		make(map[string]*NamespaceConfig)}
}

func NewDefaultNamespaceConfig() *NamespaceConfig {
	return &NamespaceConfig{Buckets: make(map[string]*BucketConfig)}
}

func NewDefaultBucketConfig() *BucketConfig {
	return &BucketConfig{Size: 100, FillRate: 50, WaitTimeoutMillis: 1000, MaxIdleMillis: -1, MaxDebtMillis: 10000}
}