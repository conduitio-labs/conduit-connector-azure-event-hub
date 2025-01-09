package azure_event_hub

import (
	"github.com/conduitio-labs/conduit-connector-azure-event-hub/destination"
	"github.com/conduitio-labs/conduit-connector-azure-event-hub/source"
	sdk "github.com/conduitio/conduit-connector-sdk"
)

// Connector combines all constructors for each plugin in one struct.
var Connector = sdk.Connector{
	NewSpecification: Specification,
	NewSource:        source.New,
	NewDestination:   destination.New,
}
