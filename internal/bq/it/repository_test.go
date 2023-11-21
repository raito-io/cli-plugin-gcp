//go:build integration

package it

import (
	"context"
	"fmt"
	"os"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/raito-io/cli/base/util/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/raito-io/cli-plugin-gcp/internal/common"
	"github.com/raito-io/cli-plugin-gcp/internal/org"
)

func TestRepository_ListDataSets(t *testing.T) {
	configMap := testSetup(t)
	ctx := context.Background()

	repository, cleanup, err := InitializeBqRepository(ctx, configMap)
	require.NoError(t, err)

	defer cleanup()

	var dataSets []org.GcpOrgEntity

	err = repository.ListDataSets(ctx, repository.Project(), func(ctx context.Context, entity *org.GcpOrgEntity, dataset *bigquery.Dataset) error {
		dataSets = append(dataSets, *entity)
		return nil
	})

	require.NoError(t, err)

	assert.Greater(t, len(dataSets), 1)
	fmt.Printf("Datasets: %+v\n", dataSets)
}

func testSetup(t *testing.T) *config.ConfigMap {
	t.Helper()

	projectId := os.Getenv(common.GcpProjectId)
	fileLocation := os.Getenv(common.GcpSAFileLocation)

	return &config.ConfigMap{
		Parameters: map[string]string{
			common.GcpProjectId:      projectId,
			common.GcpSAFileLocation: fileLocation,
		},
	}
}
