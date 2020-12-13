/*
Copyright 2019-2020 vChain, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import "github.com/codenotary/immudb/pkg/api/schema"

// Cache the cache interface
type Cache interface {
	Get(serverUuid string, dbName string) (*schema.ImmutableState, error)
	Set(root *schema.ImmutableState, serverUuid string, dbName string) error
}

// HistoryCache the history cache interface
type HistoryCache interface {
	Cache
	Walk(serverID string, dbName string, f func(*schema.ImmutableState) interface{}) ([]interface{}, error)
}
