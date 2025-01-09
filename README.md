# Conduit Connector for Azure EventHub

Azure EventHub connector is one of [Conduit](https://github.com/ConduitIO/conduit) plugins. It provides both, a Source and a Destination Azure EventHub connectors.

## How to build?
Run `make build` to build the connector.

## Testing
Run `make test` to run all the unit tests. Run `make test-integration` to run the integration tests.

The Docker compose file at `test/docker-compose.yml` can be used to run the required resource locally.

## Source

A source connector pulls data from an Azure EventHub and pushes it to downstream resources via Conduit.

### Configuration

| name                  | description                           | required | default value |
|-----------------------|---------------------------------------|----------|---------------|
| `azure.tenantId` | AzureTenantID is the azure tenant ID. | true     |           |
| `azure.clientId` | AzureClientID is the azure client ID. | true     |           |
| `azure.clientSecret` | AzureClientSecret is the azure client secret. | true     |           |
| `eventHubNamespace` | EventHubNameSpace is the FQNS of your event hubs resource. | true     |           |
| `eventHubName` | EventHubName is your event hub name- analogous to a kafka topic. | true     |           |

## Destination

A destination connector pushes data from upstream resources to Azure EventHub via Conduit.

### Configuration


| name                  | description                           | required | default value |
|-----------------------|---------------------------------------|----------|---------------|
| `azure.tenantId` | AzureTenantID is the azure tenant ID. | true     |           |
| `azure.clientId` | AzureClientID is the azure client ID. | true     |           |
| `azure.clientSecret` | AzureClientSecret is the azure client secret. | true     |           |
| `eventHubNamespace` | EventHubNameSpace is the FQNS of your event hubs resource. | true     |           |
| `eventHubName` | EventHubName is your event hub name- analogous to a kafka topic. | true     |           |
| `batchSamePartition` | BatchSamePartition is a boolean that controls whether to write batches of records to the same partition or to divide them among available partitions according to the azure event hubs producer client. | true     |      false     |


## Known Issues & Limitations

## Planned work

![scarf pixel](https://static.scarf.sh/a.png?x-pxid=)