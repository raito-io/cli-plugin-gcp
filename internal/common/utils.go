package common

import (
	"errors"
	"fmt"
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

func IsGoogle400Error(err error) bool {
	var apiError *googleapi.Error
	if !errors.As(err, &apiError) {
		return false
	}

	if apiError.Code >= 400 && apiError.Code < 500 && apiError.Code != 403 {
		Logger.Debug(fmt.Sprintf("Google 400 error: {Code :%d, Message: %s, Details: %+v, Body: %+v, Errors: %+v, err: %s}", apiError.Code, apiError.Message, apiError.Details, apiError.Body, apiError.Errors, apiError.Error()))

		return true
	}

	return false
}

func IsGoogle403Error(err error) bool {
	var apiError *googleapi.Error
	if !errors.As(err, &apiError) {
		return false
	}

	if apiError.Code == 403 {
		return true
	}

	return false
}
