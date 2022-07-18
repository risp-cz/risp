package dump

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func DecodeConfigYAML(data []byte) (configYAML *ConfigYAML, err error) {
	var rispYAML *RispYAML

	if err = yaml.Unmarshal(data, rispYAML); err != nil {
		return
	}

	if rispYAML.Config == nil {
		err = fmt.Errorf("missing config")
		return
	}

	configYAML = rispYAML.Config
	return
}

func DecodeDataYAML(data []byte) (dataYAML *DataYAML, err error) {
	var rispYAML *RispYAML

	if err = yaml.Unmarshal(data, rispYAML); err != nil {
		return
	}

	if rispYAML.Config == nil {
		err = fmt.Errorf("missing data")
		return
	}

	dataYAML = rispYAML.Data
	return
}
