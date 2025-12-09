package utils

import (
	jsoniter "github.com/json-iterator/go"
)

var (
	gJson = jsoniter.ConfigCompatibleWithStandardLibrary
)

func JsonLoads(jsonStr *string, parsedVal interface{}) error {
	jsonByteArr := []byte(*jsonStr)
	return gJson.Unmarshal(jsonByteArr, parsedVal)
}

func JsonDumps(parsedVal interface{}) (string, error) {
	s, err := gJson.Marshal(parsedVal)
	return string(s), err
}
