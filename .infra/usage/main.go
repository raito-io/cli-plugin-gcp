package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand/v2"
	"os"
	"strings"

	"cloud.google.com/go/bigquery"
	"github.com/hashicorp/go-hclog"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
)

var logger hclog.Logger

const (
	GcpSAFileLocation = "GOOGLE_APPLICATION_CREDENTIALS"

	ScopeCloudPlatform         = "https://www.googleapis.com/auth/cloud-platform"
	ScopeCloudPlatformReadOnly = "https://www.googleapis.com/auth/cloud-platform.read-only"
)

type UsageConfig struct {
	Personas struct {
		Value []string `json:"value"`
	} `json:"personas"`
	Project struct {
		Value string `json:"value"`
	} `json:"project"`
	Dataset struct {
		Value string `json:"value"`
	} `json:"dataset"`
	Tables struct {
		Value []struct {
			Dataset string   `json:"dataset"`
			Tables  []string `json:"tables"`
		} `json:"value"`
	} `json:"tables"`
}

func getConfig(scopes ...string) (*jwt.Config, error) {
	key := os.Getenv(GcpSAFileLocation)
	if key == "" {
		key = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}

	serviceAccountJSON, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, fmt.Errorf("error opening service account file: %s", err.Error())
	}

	return google.JWTConfigFromJSON(serviceAccountJSON, scopes...)
}

func connectToBigQuery(ctx context.Context, usageConfig *UsageConfig, targetPrincipal *string, subject *string) (*bigquery.Client, func(), error) {
	gcpProjectId := usageConfig.Project.Value

	config, err := getConfig(ScopeCloudPlatform)
	if err != nil {
		return nil, nil, err
	}

	if targetPrincipal != nil {
		// API needs to be enabled: https://console.cloud.google.com/apis/library/iamcredentials.googleapis.com
		ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
			TargetPrincipal: *targetPrincipal,
			Scopes:          []string{ScopeCloudPlatform},
			// Optionally supply delegates.
			// Delegates: []string{"ci-user@bq-demodata.iam.gserviceaccount.com"},
		}, option.WithHTTPClient(config.Client(ctx)))

		if err != nil {
			return nil, nil, err
		}

		logger.Info(fmt.Sprintf("Creating client and impersonating service account %s", *targetPrincipal))
		client, err := bigquery.NewClient(ctx, gcpProjectId, option.WithTokenSource(ts))
		if err != nil {
			return nil, nil, fmt.Errorf("error creating BigQuery client: %s", err.Error())
		}
		return client, func() {
			client.Close()
		}, nil
	}

	if subject != nil {
		logger.Info(fmt.Sprintf("Impersonating user %s", *subject))
		config.Scopes = []string{ScopeCloudPlatformReadOnly}
		config.Subject = *subject
	}

	logger.Info("Creating client")
	client, err := bigquery.NewClient(ctx, gcpProjectId, option.WithHTTPClient(config.Client(ctx)))
	if err != nil {
		return nil, nil, err
	}

	return client, func() { client.Close() }, nil
}

func GenerateBigQueryUsage(usageConfig *UsageConfig) error {
	userList := usageConfig.Personas.Value

	allDatasets := []string{usageConfig.Dataset.Value}

	datasetsWithTables, err := getDatasetsWithTablesFromGCP(usageConfig, allDatasets)
	if err != nil {
		return fmt.Errorf("datasets with tables: %s", err.Error())
	}

	logger.Info(fmt.Sprintf("datasets: %v", datasetsWithTables))

	for _, emailAddress := range userList {
		logger.Info(fmt.Sprintf("Executing queries for %q", emailAddress))

		var targetPrincipal, subject *string

		if strings.Contains(emailAddress, "iam.gserviceaccount.com") {
			targetPrincipal = &emailAddress
		} else {
			subject = &emailAddress
		}

		client, closeFn, err := connectToBigQuery(context.Background(), usageConfig, targetPrincipal, subject)
		if err != nil {
			return fmt.Errorf("error connecting to bigquery: %s", err.Error())
		}
		defer closeFn()

		// find out which tables a user/service account should have access to, based on all existing tables
		datasetTables := make([]DatasetTables, 0, len(usageConfig.Tables.Value))

		for i := range usageConfig.Tables.Value {
			datasetTables = append(datasetTables, DatasetTables{
				Dataset: usageConfig.Tables.Value[i].Dataset,
				Tables:  usageConfig.Tables.Value[i].Tables,
			})
		}

		queryableTables, _ := resolveAccessibleTables(datasetTables, datasetsWithTables)

		logger.Info(fmt.Sprintf("Generating usage for %s, available tables: %v", emailAddress, queryableTables))

		bytes := []byte(emailAddress)
		seed := int64(len(queryableTables))
		for _, number := range bytes {
			seed += int64(number)
		}

		for _, queryableTable := range queryableTables {
			p := rand.Float64() //Only query 6/10 tables
			if p < 0.4 {
				continue
			}

			numQueries := rand.IntN(5)
			queryString := fmt.Sprintf("SELECT * FROM %s.%s LIMIT %d", usageConfig.Project.Value, queryableTable, rand.IntN(10)+1)

			for i := 0; i < numQueries; i++ {
				logger.Info(fmt.Sprintf("Executing query for SA %s: %s", emailAddress, queryString))

				query := client.Query(queryString)
				_, err := query.Read(context.Background())
				if err != nil {
					logger.Error(fmt.Sprintf("Error executing query: %s", err.Error()))
				}

			}
		}
	}

	return nil
}

func main() {
	logger = hclog.New(&hclog.LoggerOptions{Name: "usage-logger", Level: hclog.Info})

	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if info.Mode()&os.ModeCharDevice != 0 {
		fmt.Println("The command is intended to work with pipes.")
		return
	}

	dec := json.NewDecoder(os.Stdin)

	usageConfig := UsageConfig{}

	err = dec.Decode(&usageConfig)
	if err != nil {
		panic(err)
	}

	err = GenerateBigQueryUsage(&usageConfig)
	if err != nil {
		panic(err)
	}
}
