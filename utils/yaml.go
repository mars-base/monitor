package utils

import (
	"os"

	"gopkg.in/yaml.v3"
)

// read yaml file and convert to struct
func YamlLoadFile(filename string, data interface{}) error {
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, data)
	if err != nil {
		return err
	}
	return nil
}

func YamlLoads(yamlStr *string, mapData *map[string]interface{}) error {
	err := yaml.Unmarshal([]byte(*yamlStr), mapData)
	return err
}

func YamlDumps(mapData *map[string]interface{}) (string, error) {
	d, err := yaml.Marshal(mapData)
	return string(d), err
}
