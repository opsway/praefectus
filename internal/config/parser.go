package config

import "github.com/BurntSushi/toml"

func ParseFile(path string) (*Config, error) {
	cfg := new(Config)
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
