// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package servicecfg

import "time"

// OktaConfig specifies configuration for the Okta service.
type OktaConfig struct {
	// Enabled turns the Okta service on or off for this process
	Enabled bool

	// APIEndpoint is the Okta API endpoint to use.
	APIEndpoint string

	// APITokenPath is the path to the Okta API token.
	APITokenPath string

	// SyncPeriod is the duration between synchronization calls.
	SyncPeriod time.Duration
}
