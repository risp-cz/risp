package dump

import (
	"gopkg.in/yaml.v3"
)

func EncodeConfigYAML(configYAML *ConfigYAML) (data []byte, err error) {
	return yaml.Marshal(&RispYAML{
		FormatVersion: "v0",
		Config:        configYAML,
	})
}

func EncodeDataYAML(dataYAML *DataYAML) (data []byte, err error) {
	return yaml.Marshal(&RispYAML{
		FormatVersion: "v0",
		Data:          dataYAML,
	})
}
