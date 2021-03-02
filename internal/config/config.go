package config

type Config struct {
	Server  ServerConfig
	Workers []string
	Timer   TimerConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type WorkersConfig struct {
	Command string
	Number  uint8
}

type TimerConfig struct {
	Command   string
	Frequency uint16
}
