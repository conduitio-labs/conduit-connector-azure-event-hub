package source

import "github.com/conduitio-labs/conduit-connector-azure-event-hub/common"

//go:generate paramgen -output=config_paramgen.go Config

type Config struct {
	common.Config
}
