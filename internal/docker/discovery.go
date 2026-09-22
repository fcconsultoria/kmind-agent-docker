package docker

import (
	"path"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/fcconsultoria/kmind-agent-docker/internal/config"
)

type Group struct{ Orchestrator, ComposeProject, ComposeService, SwarmStack, SwarmService string }

func Eligible(item container.Summary, filters config.Config) bool {
	name := strings.TrimPrefix(first(item.Names), "/")
	if name == "kmind-agent-docker" || item.Labels["kmind.monitoring"] == "disabled" {
		return false
	}
	if matches(item, name, filters.Containers.Exclude.Names, filters.Containers.Exclude.Images, filters.Containers.Exclude.Labels) {
		return false
	}
	include := filters.Containers.Include
	if len(include.Names) == 0 && len(include.Images) == 0 && len(include.Labels) == 0 {
		return true
	}
	return matches(item, name, include.Names, include.Images, include.Labels)
}
func GroupFor(labels map[string]string) Group {
	if service := labels["com.docker.swarm.service.name"]; service != "" {
		return Group{Orchestrator: "swarm", SwarmStack: labels["com.docker.stack.namespace"], SwarmService: service}
	}
	if project := labels["com.docker.compose.project"]; project != "" {
		return Group{Orchestrator: "compose", ComposeProject: project, ComposeService: labels["com.docker.compose.service"]}
	}
	return Group{Orchestrator: "standalone"}
}
func matches(item container.Summary, name string, names, images []string, labels map[string]string) bool {
	for _, pattern := range names {
		if ok, _ := path.Match(pattern, name); ok {
			return true
		}
	}
	for _, pattern := range images {
		if ok, _ := path.Match(pattern, item.Image); ok {
			return true
		}
	}
	for key, value := range labels {
		if item.Labels[key] == value {
			return true
		}
	}
	return false
}
func first(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
