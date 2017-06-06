package envcfg

type configFile interface {
	Contains(key string) bool
	GroupSeparator() string
	Unmarshal(filepath string, data interface{}) error
}

func newConfigFile() configFile {
	//TODO: decect file type
	return &tomlConfig{}
}
