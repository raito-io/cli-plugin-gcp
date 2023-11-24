package common

import (
	"errors"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/raito-io/cli/base"
	"google.golang.org/api/googleapi"
)

const CONTEXT_TIMEOUT = 10 * time.Second

var Logger hclog.Logger

func init() {
	Logger = base.Logger()
}

func HandleApiError(err error) bool {
	var apiError *googleapi.Error
	if !errors.As(err, &apiError) {
		return false
	}

	if apiError.Code >= 400 && apiError.Code < 500 {
		return true
	}

	return false
}
