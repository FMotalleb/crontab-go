package parser

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/FMotalleb/crontab-go/config"
)

func generateYamlFromCfg(finalConfig *config.Config) (string, error) {
	str, err := json.Marshal(finalConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal(json) final config: %v", err)
	}
	hashMap := make(map[string]any)
	if err := json.Unmarshal(str, &hashMap); err != nil {
		return "", fmt.Errorf("failed to unmarshal(json) final config: %v", err)
	}
	ans, err := yaml.Marshal(hashMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal(yaml) final config: %v", err)
	}
	result := string(ans)
	return result, nil
}
