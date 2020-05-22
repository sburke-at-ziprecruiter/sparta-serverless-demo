package config

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/sirupsen/logrus"
)

const (
	defaultConfigPath = "config.json"
	envConfigPath     = "CONFIG"
	envDDBEndpoint    = "AWS_DDB_ENDPOINT"
)

var configPath = defaultConfigPath

var Config struct {
	// These fields come from config.json
	APIName  string `json:"api_name"`  // AWS API Gateway Name
	APIStage string `json:"api_stage"` // AWS API Gateway Stage
	Bucket   string // AWS S3 Bucket Name
	Region   string // AWS Region
	Role     string // AWS IAM Role for Lambda functions
	Table    string // aWS DynamoDB Table Name

	AWS *session.Session
	DDB *dynamodb.DynamoDB
}

func init() {
	err := Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func Init() error {
	if v := os.Getenv(envConfigPath); v != "" {
		configPath = v
	}
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("Reading %s - %w", configPath, err)
	}
	err = json.Unmarshal(raw, &Config)
	if err != nil {
		return fmt.Errorf("Unmarshalling %s - %w", configPath, err)
	}
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	Config.AWS = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	// The service Endpoint can be specified in an environment variable.
	// This enables you to operate on the local DynamoDB, but obviously
	// this is of no use for Lambda functions.
	cnf := aws.NewConfig()
	if endpoint := os.Getenv(envDDBEndpoint); endpoint != "" {
		cnf = cnf.WithEndpoint(endpoint)
	}
	Config.DDB = dynamodb.New(Config.AWS, cnf)

	return nil
}

// ArchiveHook adds config.json to the Lambda zip archive.
func ArchiveHook(context map[string]interface{},
	serviceName string,
	zipWriter *zip.Writer,
	awsSession *session.Session,
	noop bool,
	logger *logrus.Logger) error {

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("Reading greeting - %w", err)
	}
	iow, err := zipWriter.Create(defaultConfigPath)
	if err != nil {
		return fmt.Errorf("Create zip/greeting - %w", err)
	}
	_, err = iow.Write(data)
	if err != nil {
		return fmt.Errorf("Writing zip/greeting - %w", err)
	}
	return nil
}
