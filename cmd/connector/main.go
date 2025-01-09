package main

import (
	sdk "github.com/conduitio/conduit-connector-sdk"

	azure_event_hub "github.com/conduitio-labs/conduit-connector-azure-event-hub"
)

func main() {
	sdk.Serve(azure_event_hub.Connector)
}
