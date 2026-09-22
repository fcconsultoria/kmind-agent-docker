package identity

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
)

type Identity struct{ ServiceInstanceID, HostID, HostName string }

func LoadOrCreate(stateDir string) (Identity, error) {
	if err := os.MkdirAll(stateDir, 0700); err != nil {
		return Identity{}, err
	}
	path := filepath.Join(stateDir, "instance-id")
	value, err := os.ReadFile(path)
	id := strings.TrimSpace(string(value))
	if err != nil || id == "" {
		raw := make([]byte, 16)
		if _, err := rand.Read(raw); err != nil {
			return Identity{}, err
		}
		id = fmt.Sprintf("%x-%x-%x-%x-%x", raw[:4], raw[4:6], raw[6:8], raw[8:10], raw[10:])
		if err := os.WriteFile(path, []byte(id+"\n"), 0600); err != nil {
			return Identity{}, err
		}
	}
	host, err := os.Hostname()
	if err != nil {
		return Identity{}, err
	}
	machine, _ := os.ReadFile("/etc/machine-id")
	return Identity{ServiceInstanceID: id, HostID: strings.TrimSpace(string(machine)), HostName: host}, nil
}
func (i Identity) ResourceAttributes(serviceName, applicationName, version string) []*commonv1.KeyValue {
	values := map[string]string{"service.name": serviceName, "service.namespace": applicationName, "service.instance.id": i.ServiceInstanceID, "host.id": i.HostID, "host.name": i.HostName, "container.runtime": "docker", "kmind.runtime.type": "docker", "kmind.agent.version": version}
	attributes := make([]*commonv1.KeyValue, 0, len(values))
	for key, value := range values {
		if value != "" {
			attributes = append(attributes, &commonv1.KeyValue{Key: key, Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: value}}})
		}
	}
	return attributes
}
