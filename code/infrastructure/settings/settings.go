package settings

import (
	"bytes"
	"fmt"
	"strings"
	"time"

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

const emptySettings = `LOCAL_ENDPOINT=
TABLE_1_ARN=
URL_SQS_ARN=
BUCKET_NAME=
RAW_PATH_PREFIX=
`

func GetSettings() (*Settings, error) {
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetDefault("CONTEXT_TIMEOUT", defaultContextTimeout)
	viper.SetDefault("LOG_TO_STDOUT", true)

	if err := viper.ReadConfig(bytes.NewBufferString(emptySettings)); err != nil {
		return nil, fmt.Errorf("error on reading empty settings: %v", err)
	}

	var settings Settings
	if err := viper.Unmarshal(&settings); err != nil {
		return nil, fmt.Errorf("error unmarshalling settings: %v", err)
	}

	// convert empty LocalEndpoint to nil
	if settings.LocalEndpoint != nil {
		if len(strings.TrimSpace(*settings.LocalEndpoint)) == 0 {
			settings.LocalEndpoint = nil
		}
	}
	return &settings, nil
}
