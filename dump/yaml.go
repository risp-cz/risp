package dump

import (
	"math"
	"strings"
)

type RispYAML struct {
	FormatVersion string      `yaml:"formatVersion"`
	Config        *ConfigYAML `yaml:"config,omitempty"`
	Data          *DataYAML   `yaml:"data,omitempty"`
}

type ConfigYAML struct {
	PathPidFile string         `yaml:"pathPidFile,omitempty"`
	PathLogFile string         `yaml:"pathLogFile,omitempty"`
	PathData    string         `yaml:"pathData,omitempty"`
	GRPC        ConfigGRPCYAML `yaml:"grpc,omitempty"`
	Repl        ConfigReplYAML `yaml:"repl,omitempty"`
}

type ConfigGRPCYAML struct {
	Host string `yaml:"host,omitempty"`
	Port int    `yaml:"port,omitempty"`
}

type ConfigReplYAML struct {
	Prompt string `yaml:"prompt,omitempty"`
}

type DataYAML struct {
	PreserveContextID bool           `yaml:"preserveContextId,omitempty"`
	PreserveSourceID  bool           `yaml:"preserveSourceId,omitempty"`
	Contexts          []*ContextYAML `yaml:"contexts,omitempty"`
	Sources           []*SourceYAML  `yaml:"sources,omitempty"`
}

type ContextYAML struct {
	Name      string        `yaml:"name,omitempty"`
	IsDefault bool          `yaml:"isDefault,omitempty"`
	Sources   []*SourceYAML `yaml:"sources,omitempty"`
}

type SourceYAML struct {
	URI       string    `yaml:"uri,omitempty"`
	Resources Resources `yaml:"resources,omitempty"`
}

type Resources []string

func (resources Resources) Len() int {
	return len(resources)
}

func (resources Resources) Less(a, b int) bool {
	aParts := strings.Split(resources[a], "/")
	bParts := strings.Split(resources[b], "/")

	if len(aParts) < len(bParts) {
		return true
	}

	partIndex := math.Min(
		float64(len(aParts)-1),
		float64(len(bParts)-1),
	)

	return aParts[int(partIndex)] < bParts[int(partIndex)]
}

func (resources Resources) Swap(a, b int) {
	resources[a], resources[b] = resources[b], resources[a]
}
