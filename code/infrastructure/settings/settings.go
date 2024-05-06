package settings

import (
	"fmt"
	"time"

	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const codeFolder = "code"
const defaultContextTimeout = 31 * time.Second

type Settings struct {
	LocalEndpoint  *string       `mapstructure:"LOCAL_ENDPOINT"`
	Table1ARN      string        `mapstructure:"TABLE_1_ARN"`
	URLSQSARN      string        `mapstructure:"URL_SQS_ARN"`
	ContextTimeout time.Duration `mapstructure:"CONTEXT_TIMEOUT"`
	DoLogToStdout  bool          `mapstructure:"LOG_TO_STDOUT"`
	BucketName     string        `mapstructure:"BUCKET_NAME"`
	RawPathPrefix  string        `mapstructure:"RAW_PATH_PREFIX"`
}

func GetSettings() (*Settings, error) {
	configPath, err := getSettingsConfigPath()
	if err != nil {
		return nil, fmt.Errorf("error on getSettingsConfigPath: %v", err)
	}
	viper.AddConfigPath(configPath)
	viper.SetConfigName("settings")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	viper.SetDefault("CONTEXT_TIMEOUT", defaultContextTimeout)
	viper.SetDefault("LOG_TO_STDOUT", true)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading in config: %v", err)
	}

	var settings Settings
	if err := viper.Unmarshal(&settings); err != nil {
		return nil, fmt.Errorf("error unmarshalling settings: %v", err)
	}
	return &settings, nil
}

func getSettingsConfigPath() (string, error) {
	initialCwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error on os.Getwd: %v", err)
	}
	cwd := initialCwd
	for filepath.Base(cwd) != codeFolder && cwd != "/" {
		cwd = filepath.Dir(cwd)
	}
	if filepath.Base(cwd) != codeFolder {
		return "", fmt.Errorf("'code' not found in initialCwd='%s'", initialCwd)
	}
	settingsPath := filepath.Join(cwd, "..")
	return settingsPath, nil
}
