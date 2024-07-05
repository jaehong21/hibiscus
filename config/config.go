package config

import "sync"

type Config struct {
	AwsProfile string
	TabKey     int
}

var (
	globalConfig *Config = nil
	configMutex  sync.RWMutex
)

func Initialize() *Config {
	configMutex.Lock()
	defer configMutex.Unlock()

	globalConfig = &Config{
		AwsProfile: "default",
		TabKey:     HOME_TAB,
	}

	return globalConfig
}

func GetConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if globalConfig == nil {
		Initialize()
	}

	return globalConfig
}

func SetAwsProfile(profile string) {
	configMutex.Lock()
	defer configMutex.Unlock()

	globalConfig.AwsProfile = profile
}

func SetTabKey(key int) {
	configMutex.Lock()
	defer configMutex.Unlock()

	globalConfig.TabKey = key
}
