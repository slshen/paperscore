package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Dir    string
	Values map[string]any
}

var config *Config

func GetConfig() *Config {
	if config != nil {
		return config
	}
	config = &Config{}
	home, err := os.UserHomeDir()
	if err != nil {
		return config
	}
	config.Dir = filepath.Join(home, ".softball")
	configFile := filepath.Join(config.Dir, "config.yaml")
	dat, err := os.ReadFile(configFile)
	if err == nil {
		if err := yaml.Unmarshal(dat, &config.Values); err != nil {
			log.Default().Printf("Cannot read config %s - %s", configFile, err)
			return config
		}
		log.Default().Printf("Loaded config from %s", configFile)
	}
	config.setFromEnviron()
	return config
}

func (config *Config) setFromEnviron() {
	for _, nv := range os.Environ() {
		eq := strings.IndexRune(nv, '=')
		n := nv[0:eq]
		if strings.HasPrefix(n, "SOFTBALL_") {
			var k strings.Builder
			cap := true
			for _, ch := range n[9:] {
				switch {
				case cap:
					k.WriteRune(unicode.ToUpper(ch))
					cap = false
				case ch == '_':
					cap = true
				default:
					k.WriteRune(unicode.ToLower(ch))
				}
			}
			config.Values[k.String()] = nv[eq+1:]
		}
	}
}

func (config *Config) Decode(s any) {
	err := mapstructure.Decode(config.Values, s)
	if err != nil {
		panic(err)
	}
}

func (config *Config) GetString(key string) string {
	val, _ := config.Values[key].(string)
	return val
}
