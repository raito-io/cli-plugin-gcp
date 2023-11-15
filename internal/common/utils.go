package common

import (
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/raito-io/cli/base"
)

const CONTEXT_TIMEOUT = 10 * time.Second

var Logger hclog.Logger

func init() {
	Logger = base.Logger()
}
