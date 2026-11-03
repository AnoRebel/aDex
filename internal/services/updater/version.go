// Package updater wires aDex-UI into the Wails in-app updater.
package updater

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseVersion extracts info.version from the contents of build/config.yml,
// the project's single source of truth for the version. The result has no
// leading "v", which is the form the Wails updater compares against release
// tags. An unparseable file or an empty version is an error: an updater that
// guessed its own version could offer a downgrade or never offer anything.
func ParseVersion(configYAML []byte) (string, error) {
	var cfg struct {
		Info struct {
			Version string `yaml:"version"`
		} `yaml:"info"`
	}
	if err := yaml.Unmarshal(configYAML, &cfg); err != nil {
		return "", fmt.Errorf("parse build config: %w", err)
	}
	v := strings.TrimPrefix(strings.TrimSpace(cfg.Info.Version), "v")
	if v == "" {
		return "", errors.New("build config has no info.version")
	}
	return v, nil
}
