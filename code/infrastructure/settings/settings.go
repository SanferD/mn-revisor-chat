package settings

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const relativeSettingsFilePath = "../../../settings.env"
const defaultContextTimeout = 59 * time.Second
const defaultEmbeddingModelID = "amazon.titan-embed-text-v2:0"
const defaultFoundationModelID = "anthropic.claude-v2"

type Settings struct {
	ContextTimeout time.Duration `mapstructure:"CONTEXT_TIMEOUT"`
	DoLogToStdout  bool          `mapstructure:"LOG_TO_STDOUT"`
	LocalEndpoint  *string       `mapstructure:"LOCAL_ENDPOINT"`
	// s3
	MainBucketName  string `mapstructure:"MAIN_BUCKET_NAME"`
	ChunkPathPrefix string `mapstructure:"CHUNK_PATH_PREFIX"`
	RawPathPrefix   string `mapstructure:"RAW_PATH_PREFIX"`
	// sqs
	URLSQSARN       string `mapstructure:"URL_SQS_ARN"`
	RawEventsSQSARN string `mapstructure:"RAW_EVENTS_SQS_ARN"`
	ToIndexSQSARN   string `mapstructure:"TO_INDEX_SQS_ARN"`
	// ddb
	Table1ARN string `mapstructure:"TABLE_1_ARN"`
	// ecs
	TriggerCrawlerTaskDfnArn string   `mapstructure:"TRIGGER_CRAWLER_TASK_DFN_ARN"`
	TriggerCrawlerClusterArn string   `mapstructure:"TRIGGER_CRAWLER_CLUSTER_ARN"`
	SubnetIds                []string `mapstructure:"PRIVATE_ISOLATED_SUBNET_IDS"`
	SecurityGroupIds         []string `mapstructure:"SECURITY_GROUP_IDS"`
	// bedrock
	EmbeddingModelID  string `mapstructure:"EMBEDDING_MODEL_ID"`
	FoundationModelID string `mapstructure:"FOUNDATION_MODEL_ID"`
	// opensearch
	OpensearchUsername        string `mapstructure:"OPENSEARCH_USERNAME"`
	OpensearchPassword        string `mapstructure:"OPENSEARCH_PASSWORD"`
	OpensearchDomain          string `mapstructure:"OPENSEARCH_DOMAIN"`
	DoAllowOpensearchInsecure bool   `mapstructure:"DO_ALLOW_OPENSEARCH_INSECURE"`
	OpensearchIndexName       string `mapstructure:"OPENSEARCH_INDEX_NAME"`
	// sinch
	SinchAPIToken           string `mapstructure:"SINCH_API_TOKEN"`
	SinchProjectID          string `mapstructure:"SINCH_PROJECT_ID"`
	SinchVirtualPhoneNumber string `mapstructure:"SINCH_VIRTUAL_PHONE_NUMBER"`
}

const emptySettings = `
LOCAL_ENDPOINT=
MAIN_BUCKET_NAME=
CHUNK_PATH_PREFIX=
RAW_PATH_PREFIX=
URL_SQS_ARN=
RAW_EVENTS_SQS_ARN=
TO_INDEX_SQS_ARN=
TABLE_1_ARN=
PRIVATE_ISOLATED_SUBNET_IDS=
SECURITY_GROUP_IDS=
OPENSEARCH_USERNAME=
OPENSEARCH_PASSWORD=
OPENSEARCH_DOMAIN=
OPENSEARCH_INDEX_NAME=
`

func GetSettings() (*Settings, error) {
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetDefault("CONTEXT_TIMEOUT", defaultContextTimeout)
	viper.SetDefault("LOG_TO_STDOUT", true)
	viper.SetDefault("TRIGGER_CRAWLER_TASK_DFN_ARN", "")
	viper.SetDefault("TRIGGER_CRAWLER_CLUSTER_ARN", "")
	viper.SetDefault("EMBEDDING_MODEL_ID", defaultEmbeddingModelID)
	viper.SetDefault("FOUNDATION_MODEL_ID", defaultFoundationModelID)
	viper.SetDefault("DO_ALLOW_OPENSEARCH_INSECURE", false)
	viper.SetDefault("SINCH_API_TOKEN", "")
	viper.SetDefault("SINCH_PROJECT_ID", "")
	viper.SetDefault("SINCH_VIRTUAL_PHONE_NUMBER", "")

	// load settings

	//// read empty settings string to "prime" it for reading from environment (otherwise doesn't work..)
	if err := viper.ReadConfig(bytes.NewBufferString(emptySettings)); err != nil {
		return nil, fmt.Errorf("error on reading empty settings: %v", err)
	}

	//// Check if settings file exists and load it
	if _, err := os.Stat(relativeSettingsFilePath); err == nil {
		// Use a relative path for the settings file based on the executable location
		viper.SetConfigFile(relativeSettingsFilePath)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("error loading settings file: %v", err)
		}
	}

	//// load settings (environment precedent over settings file)
	var settings Settings
	if err := viper.Unmarshal(&settings); err != nil {
		return nil, fmt.Errorf("error unmarshalling settings: %v", err)
	}

	// sanitize settings

	//// convert empty LocalEndpoint to nil
	if settings.LocalEndpoint != nil {
		if len(strings.TrimSpace(*settings.LocalEndpoint)) == 0 {
			settings.LocalEndpoint = nil
		}
	}
	log.Printf("%+v\n", settings)
	return &settings, nil
}
