package docker

import (
	"github.com/docker/docker/api/types/container"
	"github.com/fcconsultoria/kmind-agent-docker/internal/config"
	"testing"
)

func TestEligibleExcludesAgentAndDisabled(t *testing.T) {
	var cfg config.Config
	if Eligible(container.Summary{Names: []string{"/kmind-agent-docker"}}, cfg) || Eligible(container.Summary{Labels: map[string]string{"kmind.monitoring": "disabled"}}, cfg) {
		t.Fatal("protected containers must be excluded")
	}
}
func TestGroupForComposeAndSwarm(t *testing.T) {
	if got := GroupFor(map[string]string{"com.docker.compose.project": "shop", "com.docker.compose.service": "api"}); got.Orchestrator != "compose" || got.ComposeService != "api" {
		t.Fatal("compose labels not grouped")
	}
	if got := GroupFor(map[string]string{"com.docker.swarm.service.name": "api"}); got.Orchestrator != "swarm" {
		t.Fatal("swarm labels not grouped")
	}
}
