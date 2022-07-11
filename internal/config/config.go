package config

import (
	"os"
	"strconv"
	"time"
)

const (
	MaxProcessPoolSizeEnv          string = "PRAEFECTUS_MAX_PROCESS_POOL_SIZE"
	WorkerBusynessLowEnv           string = "PRAEFECTUS_WORKER_BUSINESS_LOW"
	WorkerBusynessAverageEnv       string = "PRAEFECTUS_WORKER_BUSINESS_AVERAGE"
	WorkerBusynessHighEnv          string = "PRAEFECTUS_WORKER_BUSINESS_HIGH"
	WorkerNumberLowIncreaseEnv     string = "PRAEFECTUS_WORKER_NUMBER_LOW_INCREASE"
	WorkerNumberAverageIncreaseEnv string = "PRAEFECTUS_WORKER_NUMBER_AVERAGE_INCREASE"
	WorkerNumberHighIncreaseEnv    string = "PRAEFECTUS_WORKER_NUMBER_HIGH_INCREASE"
	WorkerIdlePercentageLimitEnv   string = "PRAEFECTUS_WORKER_IDLE_PERCENTAGE_LIMIT"
	ScaleTickEnv                   string = "PRAEFECTUS_POOL_SCALE_TICK"
	DownscaleTickEnv               string = "PRAEFECTUS_POOL_DOWNSCALE_TICK"
)

const (
	MaxProcessPullSize          uint8         = 5
	WorkerBusynessLow                         = uint8(50)
	WorkerBusynessAverage                     = uint8(70)
	WorkerBusynessHigh                        = uint8(90)
	WorkerNumberLowIncrease                   = uint8(1)
	WorkerNumberAverageIncrease               = uint8(3)
	WorkerNumberHighIncrease                  = uint8(5)
	WorkerIdlePercentageLimit                 = uint8(80)
	ScaleTick                   time.Duration = 30
	DownscaleTick               time.Duration = 30
)

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

type ScalePoolConfig struct {
	WorkerBusynessLow           uint8
	WorkerBusynessAverage       uint8
	WorkerBusynessHigh          uint8
	WorkerNumberLowIncrease     uint8
	WorkerNumberAverageIncrease uint8
	WorkerNumberHighIncrease    uint8
	WorkerIdlePercentageLimit   uint8
	ScaleTick                   time.Duration
	DownscaleTick               time.Duration
	MaxProcessPullSize          uint8
}

func SetupPoolConfig() *ScalePoolConfig {
	poolConfig := &ScalePoolConfig{
		MaxProcessPullSize:          MaxProcessPullSize,
		WorkerBusynessLow:           WorkerBusynessLow,
		WorkerBusynessAverage:       WorkerBusynessAverage,
		WorkerBusynessHigh:          WorkerBusynessHigh,
		WorkerNumberLowIncrease:     WorkerNumberLowIncrease,
		WorkerNumberAverageIncrease: WorkerNumberAverageIncrease,
		WorkerNumberHighIncrease:    WorkerNumberHighIncrease,
		WorkerIdlePercentageLimit:   WorkerIdlePercentageLimit,
		ScaleTick:                   ScaleTick,
		DownscaleTick:               DownscaleTick,
	}

	envBufferSize, found := parseUint8Env(MaxProcessPoolSizeEnv)
	if found {
		poolConfig.MaxProcessPullSize = envBufferSize
	}

	envPercentageLimit, found := parseUint8Env(WorkerIdlePercentageLimitEnv)
	if found {
		poolConfig.WorkerIdlePercentageLimit = envPercentageLimit
	}

	envWorkerBusynessLow, found := parseUint8Env(WorkerBusynessLowEnv)
	if found {
		poolConfig.WorkerBusynessLow = envWorkerBusynessLow
	}

	envWorkerBusynessAverage, found := parseUint8Env(WorkerBusynessAverageEnv)
	if found {
		poolConfig.WorkerBusynessAverage = envWorkerBusynessAverage
	}

	envWorkerBusynessHigh, found := parseUint8Env(WorkerBusynessHighEnv)
	if found {
		poolConfig.WorkerBusynessHigh = envWorkerBusynessHigh
	}

	envWorkerLowIncrease, found := parseUint8Env(WorkerNumberLowIncreaseEnv)
	if found {
		poolConfig.WorkerNumberLowIncrease = envWorkerLowIncrease
	}

	envWorkerAverageIncrease, found := parseUint8Env(WorkerNumberAverageIncreaseEnv)
	if found {
		poolConfig.WorkerNumberAverageIncrease = envWorkerAverageIncrease
	}

	envWorkerHighIncrease, found := parseUint8Env(WorkerNumberHighIncreaseEnv)
	if found {
		poolConfig.WorkerNumberHighIncrease = envWorkerHighIncrease
	}

	envScaleTick, found := parseDurationEnv(ScaleTickEnv)
	if found {
		poolConfig.ScaleTick = envScaleTick
	}

	envDownscaleTick, found := parseDurationEnv(DownscaleTickEnv)
	if found {
		poolConfig.DownscaleTick = envDownscaleTick
	}

	return poolConfig
}

func parseUint8Env(env string) (uint8, bool) {
	envValue, found := os.LookupEnv(env)
	if found {
		parsedValue, err := strconv.ParseUint(envValue, 10, 8)
		if err == nil {
			return uint8(parsedValue), true
		}
	}
	return uint8(0), false
}

func parseDurationEnv(env string) (time.Duration, bool) {
	envValue, found := os.LookupEnv(env)
	if found {
		parsedValue, err := strconv.ParseUint(envValue, 10, 64)
		if err == nil {
			return time.Duration(parsedValue), true
		}
	}
	return time.Duration(0), false
}
