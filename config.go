package envcfg

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type configFile interface {
	Contains(key string) bool
	GroupSeparator() string
	Unmarshal(filepath string, data interface{}) error
}

func newConfigFile() configFile {
	//TODO: decect file type
	return &tomlConfig{}
}

func configFilePath(path, ext string) (string, error) {
	var parent string = ""
	if path != "" {
		stat, err := os.Stat(path)
		if err == nil {
			if stat.IsDir() {
				parent = path
			} else {
				//all good
				return path, nil
			}
		}
	}
	//search config file
	if parent == "" {
		var err error
		//set executable parent
		parent, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return "", err
		}
	}
	//search file in parent with some extension
	files, err := ioutil.ReadDir(parent)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ext) {
			return f.Name(), nil
		}
	}
	return "", errors.New("config file not found")
}
