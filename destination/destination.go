// Copyright © 2023 Meroxa, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package destination

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"github.com/conduitio/conduit-commons/config"
	"github.com/conduitio/conduit-commons/opencdc"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/oklog/ulid/v2"
)

type Destination struct {
	sdk.UnimplementedDestination

	config Config
	client *azeventhubs.ProducerClient
}

func New() sdk.Destination {
	return sdk.DestinationWithMiddleware(&Destination{}, sdk.DefaultDestinationMiddleware()...)
}

func (d *Destination) Parameters() config.Parameters {
	return Config{}.Parameters()
}

func (d *Destination) Configure(ctx context.Context, cfg config.Config) error {
	sdk.Logger(ctx).Info().Msg("Configuring Destination...")
	err := sdk.Util.ParseConfig(ctx, cfg, &d.config, New().Parameters())
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	return nil
}

func (d *Destination) Open(ctx context.Context) error {
	defaultAzureCred, err := azidentity.NewClientSecretCredential(d.config.AzureTenantID, d.config.AzureClientID, d.config.AzureClientSecret, nil)
	if err != nil {
		return err
	}

	d.client, err = azeventhubs.NewProducerClient(d.config.EventHubNameSpace, d.config.EventHubName, defaultAzureCred, nil)
	if err != nil {
		return err
	}

	return nil
}

func (d *Destination) Write(ctx context.Context, records []opencdc.Record) (int, error) {
	var written int
	var newBatchOptions *azeventhubs.EventDataBatchOptions
	if d.config.BatchSamePartition {
		partitionKey, err := ulid.New(ulid.Now(), nil)
		if err != nil {
			return 0, err
		}
		newBatchOptions = &azeventhubs.EventDataBatchOptions{
			PartitionKey: to.Ptr[string](partitionKey.String()),
		}
	}

	batch, err := d.client.NewEventDataBatch(ctx, newBatchOptions)
	if err != nil {
		return 0, err
	}

	for i := 0; i < len(records); i++ {
		err := batch.AddEventData(&azeventhubs.EventData{
			Body:        records[i].Bytes(),
			ContentType: to.Ptr[string]("application/json"),
		}, nil)
		if err != nil {
			if errors.Is(err, azeventhubs.ErrEventDataTooLarge) {
				if batch.NumEvents() == 0 {
					// record is too large to write to destination, log error and continue
					sdk.Logger(ctx).Error().Msgf("record with key %s is too large to be sent, wasn't written", records[i].Key.Bytes())
					return written, err
				}

				// batch is full, send and add the current event to the next batch by decrementing the iterator
				if err := d.client.SendEventDataBatch(ctx, batch, nil); err != nil {
					sdk.Logger(ctx).Err(err)
					return written, err
				}

				written += int(batch.NumEvents())

				tmpBatch, err := d.client.NewEventDataBatch(ctx, newBatchOptions)
				if err != nil {
					sdk.Logger(ctx).Err(err)
					return written, err
				}

				batch = tmpBatch
				i--
			} else {
				return written, err
			}
		}
	}

	// write any remaining events in the batch
	if batch.NumEvents() > 0 {
		if err := d.client.SendEventDataBatch(ctx, batch, nil); err != nil {
			return written, err
		}

		written += int(batch.NumEvents())
	}

	return written, nil
}

func (d *Destination) Teardown(ctx context.Context) error {
	if d.client != nil {
		err := d.client.Close(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
