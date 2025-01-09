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

package common

type Config struct {
	// AzureTenantID is the azure tenant ID
	AzureTenantID string `json:"azure.tenantId" validate:"required"`
	// AzureClientID is the azure client ID
	AzureClientID string `json:"azure.clientId" validate:"required"`
	// AzureClientSecret is the azure client secret
	AzureClientSecret string `json:"azure.clientSecret" validate:"required"`

	// EventHubNameSpace is the FQNS of your event hubs resource
	EventHubNameSpace string `json:"eventHubNamespace" validate:"required"`
	// EventHubName is your event hub name- analogous to a kafka topic
	EventHubName string `json:"eventHubName" validate:"required"`
}
