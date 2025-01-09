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

package source

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
	"github.com/oklog/ulid/v2"
)

type checkpointStore struct {
	checkpointsMu sync.RWMutex
	checkpoints   map[string]azeventhubs.Checkpoint

	ownershipMu sync.RWMutex
	ownerships  map[string]azeventhubs.Ownership
}

func newCheckpointStore() *checkpointStore {
	return &checkpointStore{
		checkpoints: map[string]azeventhubs.Checkpoint{},
		ownerships:  map[string]azeventhubs.Ownership{},
	}
}

func (cps *checkpointStore) ExpireOwnership(o azeventhubs.Ownership) {
	key := strings.Join([]string{o.FullyQualifiedNamespace, o.EventHubName, o.ConsumerGroup, o.PartitionID}, "/")

	cps.ownershipMu.Lock()
	defer cps.ownershipMu.Unlock()

	oldO := cps.ownerships[key]
	oldO.LastModifiedTime = time.Now().UTC().Add(-2 * time.Hour)
	cps.ownerships[key] = oldO
}

func (cps *checkpointStore) ReqlinquishOwnership(o azeventhubs.Ownership) {
	key := strings.Join([]string{o.FullyQualifiedNamespace, o.EventHubName, o.ConsumerGroup, o.PartitionID}, "/")

	cps.ownershipMu.Lock()
	defer cps.ownershipMu.Unlock()

	oldO := cps.ownerships[key]
	oldO.OwnerID = ""
	cps.ownerships[key] = oldO
}

func (cps *checkpointStore) ClaimOwnership(ctx context.Context, partitionOwnership []azeventhubs.Ownership, options *azeventhubs.ClaimOwnershipOptions) ([]azeventhubs.Ownership, error) {
	var owned []azeventhubs.Ownership

	for _, po := range partitionOwnership {
		ownership, err := func(po azeventhubs.Ownership) (*azeventhubs.Ownership, error) {
			cps.ownershipMu.Lock()
			defer cps.ownershipMu.Unlock()

			if po.ConsumerGroup == "" ||
				po.EventHubName == "" ||
				po.FullyQualifiedNamespace == "" ||
				po.PartitionID == "" {
				panic("bad test, not all required fields were filled out for ownership data")
			}

			key := strings.Join([]string{po.FullyQualifiedNamespace, po.EventHubName, po.ConsumerGroup, po.PartitionID}, "/")

			current, exists := cps.ownerships[key]

			if exists {
				if po.ETag == nil {
					panic("Ownership blob exists, we should have claimed it using an etag")
				}

				if *po.ETag != *current.ETag {
					// can't own it, didn't have the expected etag
					return nil, nil
				}
			}

			newOwnership := po
			uuid, err := ulid.New(ulid.Now(), nil)
			if err != nil {
				return nil, err
			}

			newOwnership.ETag = to.Ptr[azcore.ETag](azcore.ETag(uuid.String()))
			newOwnership.LastModifiedTime = time.Now().UTC()
			cps.ownerships[key] = newOwnership

			return &newOwnership, nil
		}(po)
		if err != nil {
			return nil, err
		}

		if ownership != nil {
			owned = append(owned, *ownership)
		}
	}

	return owned, nil
}

func (cps *checkpointStore) ListCheckpoints(ctx context.Context, fullyQualifiedNamespace string, eventHubName string, consumerGroup string, options *azeventhubs.ListCheckpointsOptions) ([]azeventhubs.Checkpoint, error) {
	cps.checkpointsMu.RLock()
	defer cps.checkpointsMu.RUnlock()

	var checkpoints []azeventhubs.Checkpoint

	for _, v := range cps.checkpoints {
		checkpoints = append(checkpoints, v)
	}

	return checkpoints, nil
}

func (cps *checkpointStore) ListOwnership(ctx context.Context, fullyQualifiedNamespace string, eventHubName string, consumerGroup string, options *azeventhubs.ListOwnershipOptions) ([]azeventhubs.Ownership, error) {
	cps.ownershipMu.RLock()
	defer cps.ownershipMu.RUnlock()

	var ownerships []azeventhubs.Ownership

	for _, v := range cps.ownerships {
		ownerships = append(ownerships, v)
	}

	sort.Slice(ownerships, func(i, j int) bool {
		return ownerships[i].PartitionID < ownerships[j].PartitionID
	})

	return ownerships, nil
}

func (cps *checkpointStore) SetCheckpoint(ctx context.Context, checkpoint azeventhubs.Checkpoint, options *azeventhubs.SetCheckpointOptions) error {
	cps.checkpointsMu.Lock()
	defer cps.checkpointsMu.Unlock()

	if checkpoint.ConsumerGroup == "" ||
		checkpoint.EventHubName == "" ||
		checkpoint.FullyQualifiedNamespace == "" ||
		checkpoint.PartitionID == "" {
		panic("bad test, not all required fields were filled out for checkpoint data")
	}

	key := toInMemoryKey(checkpoint)
	cps.checkpoints[key] = checkpoint

	return nil
}

func toInMemoryKey(a azeventhubs.Checkpoint) string {
	return strings.Join([]string{a.FullyQualifiedNamespace, a.EventHubName, a.ConsumerGroup, a.PartitionID}, "/")
}
