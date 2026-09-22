package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesEnvironmentKeyBeforeYAMLKeyFile(t *testing.T) {
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key")
	if err := os.WriteFile(keyFile, []byte("kmd_file_key_1234567890\n"), 0600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "config.yaml")
	content := "kmind:\n  apiKeyFile: '" + keyFile + "'\n  serviceName: docker-prod\n  applicationName: infrastructure\n  endpoint: https://ingest.kmind.com.br\n  statusEndpoint: https://ingest.kmind.com.br/v1/agent/status\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("KMIND_API_KEY", "kmd_environment_key_1234567890")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Kmind.APIKey != "kmd_environment_key_1234567890" {
		t.Fatalf("API key = %q, want environment value", cfg.Kmind.APIKey)
	}
}

func TestLoadRejectsUnsafeMetricInterval(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := "kmind:\n  apiKey: kmd_valid_key_1234567890\n  serviceName: docker-prod\n  applicationName: infrastructure\n  endpoint: https://ingest.kmind.com.br\n  statusEndpoint: https://ingest.kmind.com.br/v1/agent/status\ncollection:\n  metricsInterval: 10s\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("expected interval validation error")
	}
}
