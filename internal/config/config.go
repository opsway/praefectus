package config

type Config struct {
	Server  ServerConfig
	Workers map[string]WorkersConfig
	Timers  map[string]TimersConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type WorkersConfig struct {
	Command string
	Number  uint8
}

type TimersConfig struct {
	Command   string
	Frequency uint16
}
