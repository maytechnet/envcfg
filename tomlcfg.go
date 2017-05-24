package envcfg

import (
	"strings"

	"github.com/BurntSushi/toml"
)

type tomlConfig struct {
	metadata toml.MetaData
}

func (t *tomlConfig) Contains(key string) bool {
	for _, v := range t.metadata.Keys() {
		if strings.EqualFold(key, v.String()) {
			return true
		}
	}
	return false
}

func (t *tomlConfig) GroupSeparator() string {
	return "."
}

func (c *tomlConfig) Unmarshal(filepath string, data interface{}) error {
	var err error
	c.metadata, err = toml.DecodeFile(filepath, data)
	if err != nil {
		return err
	}
	return nil
}
