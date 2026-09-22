package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultInterval = 90 * time.Second
	minimumInterval = 30 * time.Second
	defaultTimeout  = 5 * time.Second
)

type Duration time.Duration

func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	value, err := time.ParseDuration(node.Value)
	if err != nil {
		return err
	}
	*d = Duration(value)
	return nil
}
func (d Duration) Value() time.Duration { return time.Duration(d) }

type Config struct {
	Kmind struct {
		APIKey          string `yaml:"apiKey"`
		APIKeyFile      string `yaml:"apiKeyFile"`
		ServiceName     string `yaml:"serviceName"`
		ApplicationName string `yaml:"applicationName"`
		Endpoint        string `yaml:"endpoint"`
		StatusEndpoint  string `yaml:"statusEndpoint"`
	} `yaml:"kmind"`
	Docker struct {
		Socket     string `yaml:"socket"`
		APIVersion string `yaml:"apiVersion"`
	} `yaml:"docker"`
	Collection struct {
		MetricsInterval         Duration `yaml:"metricsInterval"`
		DiskUsageInterval       Duration `yaml:"diskUsageInterval"`
		MetadataRefreshInterval Duration `yaml:"metadataRefreshInterval"`
	} `yaml:"collection"`
	Containers struct {
		Include struct {
			Names  []string          `yaml:"names"`
			Images []string          `yaml:"images"`
			Labels map[string]string `yaml:"labels"`
		} `yaml:"include"`
		Exclude struct {
			Names  []string          `yaml:"names"`
			Images []string          `yaml:"images"`
			Labels map[string]string `yaml:"labels"`
		} `yaml:"exclude"`
	} `yaml:"containers"`
	Limits struct {
		MemoryBufferBytes    int      `yaml:"memoryBufferBytes"`
		MaxMemoryBufferBytes int      `yaml:"maxMemoryBufferBytes"`
		MaxBatchBytes        int      `yaml:"maxBatchBytes"`
		RequestTimeout       Duration `yaml:"requestTimeout"`
	} `yaml:"limits"`
	StateDir string `yaml:"stateDir"`
}

func Load(path string) (Config, error) {
	var cfg Config
	if path != "" {
		if raw, err := os.ReadFile(path); err == nil {
			if err := yaml.Unmarshal(raw, &cfg); err != nil {
				return cfg, fmt.Errorf("parse configuration: %w", err)
			}
		} else if !os.IsNotExist(err) {
			return cfg, fmt.Errorf("read configuration: %w", err)
		}
	}
	applyEnvironment(&cfg)
	// An explicit environment key wins over both YAML apiKey and apiKeyFile.
	if cfg.Kmind.APIKey == "" && cfg.Kmind.APIKeyFile != "" {
		raw, err := os.ReadFile(cfg.Kmind.APIKeyFile)
		if err != nil {
			return cfg, fmt.Errorf("read kmind API key file: %w", err)
		}
		cfg.Kmind.APIKey = strings.TrimSpace(string(raw))
	}
	applyDefaults(&cfg)
	return cfg, cfg.Validate()
}

func applyEnvironment(c *Config) {
	set := func(key string, dst *string) {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			*dst = value
		}
	}
	set("KMIND_API_KEY", &c.Kmind.APIKey)
	set("KMIND_API_KEY_FILE", &c.Kmind.APIKeyFile)
	set("KMIND_SERVICE_NAME", &c.Kmind.ServiceName)
	set("KMIND_APPLICATION_NAME", &c.Kmind.ApplicationName)
	set("KMIND_ENDPOINT", &c.Kmind.Endpoint)
	set("KMIND_STATUS_ENDPOINT", &c.Kmind.StatusEndpoint)
	set("DOCKER_HOST", &c.Docker.Socket)
	parse := func(key string, dst *Duration) {
		if raw := os.Getenv(key); raw != "" {
			if value, err := time.ParseDuration(raw); err == nil {
				*dst = Duration(value)
			}
		}
	}
	parse("KMIND_COLLECTION_INTERVAL", &c.Collection.MetricsInterval)
	parse("KMIND_DISK_USAGE_INTERVAL", &c.Collection.DiskUsageInterval)
}
func applyDefaults(c *Config) {
	if c.Docker.Socket == "" {
		c.Docker.Socket = "unix:///var/run/docker.sock"
	}
	if c.Collection.MetricsInterval == 0 {
		c.Collection.MetricsInterval = Duration(defaultInterval)
	}
	if c.Collection.DiskUsageInterval == 0 {
		c.Collection.DiskUsageInterval = Duration(15 * time.Minute)
	}
	if c.Collection.MetadataRefreshInterval == 0 {
		c.Collection.MetadataRefreshInterval = Duration(5 * time.Minute)
	}
	if c.Limits.RequestTimeout == 0 {
		c.Limits.RequestTimeout = Duration(defaultTimeout)
	}
	if c.Limits.MemoryBufferBytes == 0 {
		c.Limits.MemoryBufferBytes = 2 * 1024 * 1024
	}
	if c.Limits.MaxMemoryBufferBytes == 0 {
		c.Limits.MaxMemoryBufferBytes = 5 * 1024 * 1024
	}
	if c.Limits.MaxBatchBytes == 0 {
		c.Limits.MaxBatchBytes = 512 * 1024
	}
	if c.StateDir == "" {
		c.StateDir = "/var/lib/kmind-agent"
	}
}
func (c Config) Validate() error {
	if !strings.HasPrefix(c.Kmind.APIKey, "kmd_") || len(c.Kmind.APIKey) < 20 {
		return fmt.Errorf("kmind API key must be a valid Docker Agent key")
	}
	if strings.TrimSpace(c.Kmind.ServiceName) == "" || strings.TrimSpace(c.Kmind.ApplicationName) == "" {
		return fmt.Errorf("kmind.serviceName and kmind.applicationName are required")
	}
	for field, value := range map[string]string{"kmind.endpoint": c.Kmind.Endpoint, "kmind.statusEndpoint": c.Kmind.StatusEndpoint} {
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("%s must be an HTTPS URL", field)
		}
	}
	if !strings.HasPrefix(c.Docker.Socket, "unix://") {
		return fmt.Errorf("docker.socket must use unix://")
	}
	if c.Collection.MetricsInterval.Value() < minimumInterval {
		return fmt.Errorf("collection.metricsInterval must be at least %s", minimumInterval)
	}
	if c.Collection.DiskUsageInterval.Value() < c.Collection.MetricsInterval.Value() || c.Collection.MetadataRefreshInterval.Value() < c.Collection.MetricsInterval.Value() {
		return fmt.Errorf("metadata and disk intervals must not be lower than metricsInterval")
	}
	if c.Limits.MemoryBufferBytes < 64*1024 || c.Limits.MaxMemoryBufferBytes < c.Limits.MemoryBufferBytes || c.Limits.MaxMemoryBufferBytes > 5*1024*1024 {
		return fmt.Errorf("invalid memory buffer limits")
	}
	if c.Limits.MaxBatchBytes < 1024 || c.Limits.MaxBatchBytes > 512*1024 || c.Limits.RequestTimeout.Value() <= 0 || c.Limits.RequestTimeout.Value() > 30*time.Second {
		return fmt.Errorf("invalid transport limits")
	}
	return nil
}
func ConfigPathFromEnv() string {
	if path := strings.TrimSpace(os.Getenv("KMIND_CONFIG_FILE")); path != "" {
		return filepath.Clean(path)
	}
	return "/etc/kmind-agent/config.yaml"
}
