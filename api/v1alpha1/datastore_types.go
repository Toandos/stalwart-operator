/*
Copyright 2026.

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

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type StalwartDataStore struct {
	// +kubebuilder:validation:Enum=RocksDb;Sqlite;FoundationDb;PostgreSql;MySql
	Type string `json:"type"`

	RocksDb      *RocksDbDataStore      `json:"rocksDb,omitempty"`
	Sqlite       *SqliteDataStore       `json:"sqlite,omitempty"`
	FoundationDb *FoundationDbDataStore `json:"foundationDb,omitempty"`
	PostgreSql   *PostgreSqlDataStore   `json:"postgresql,omitempty"`
	MySql        *MySqlDataStore        `json:"mysql,omitempty"`
}

type RocksDbDataStore struct {
	// Path to the RocksDB data directory
	// +required
	Path string `json:"path"`

	// Minimum size of a blob to store in the blob store, smaller blobs are stored in the metadata store
	// +kubebuilder:validation:Minimum=1024
	// +kubebuilder:validation:Maximum=1048576
	BlobSize int `json:"blobSize,omitempty"`

	// Total amount of memory shared by the write buffers (memtables) of every column family. Higher values batch more writes in memory before they are flushed to disk
	// +kubebuilder:validation:Minimum=8388608
	// +kubebuilder:validation:Maximum=4294967296
	BufferSize int `json:"bufferSize,omitempty"`

	// Number of worker threads to use for the store, defaults to the number of cores
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=64
	PoolWorkers int `json:"poolWorkers,omitempty"`

	// Total amount of memory shared by the block caches of every column family, used to cache data blocks, indexes and bloom filters read from disk
	// +kubebuilder:validation:Minimum=8388608
	// +kubebuilder:validation:Maximum=17179869184
	CacheSize int `json:"cacheSize,omitempty"`
}

type SqliteDataStore struct {
	// Path to the SQLite data directory
	// +required
	Path string `json:"path"`

	// Number of worker threads to use for the store, defaults to the number of cores
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=64
	PoolWorkers int `json:"poolWorkers,omitempty"`

	// Maximum number of connections to the store
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=8192
	PoolMaxConnections int `json:"poolMaxConnections,omitempty"`
}

type FoundationDbDataStore struct {
	// Path to the cluster file for the FoundationDB cluster
	ClusterFile string `json:"clusterFile,omitempty"`

	// Data center ID
	DatacenterId string `json:"datacenterId,omitempty"`

	// Machine ID in the FoundationDB cluster
	MachineId string `json:"machineId,omitempty"`

	// Transaction maximum retry delay
	TransactionRetryDelay int `json:"transactionRetryDelay,omitempty"`

	// Transaction maximum retry delay
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=1000
	TransactionRetryLimit int `json:"transactionRetryLimit,omitempty"`

	// Transaction timeout
	TransactionTimeout int `json:"transactionTimeout,omitempty"`
}

type PostgreSqlDataStore struct {
	// Connection timeout to the database
	Timeout int `json:"timeout,omitempty"`

	// Use TLS to connect to the store
	UseTls bool `json:"useTls,omitempty"`

	// Allow invalid TLS certificates when connecting to the store
	AllowInvalidCerts bool `json:"allowInvalidCerts,omitempty"`

	// Maximum number of connections to the store
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=8192
	PoolMaxConnections int `json:"poolMaxConnections,omitempty"`

	// Method to use when recycling connections in the pool
	PoolRecyclingMethod string `json:"poolRecyclingMethod,omitempty"`

	// Hostname of the database server
	// +required
	Host string `json:"host"`

	// Port of the database server
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port,omitempty"`

	// Name of the database
	Database string `json:"database,omitempty"`

	// Username to connect to the store
	AuthUsername string `json:"authUsername,omitempty"`

	// Password to connect to the store
	// +required
	AuthSecret string `json:"authSecret"`

	// Additional connection options
	Options string `json:"options,omitempty"`
}

type MySqlDataStore struct {
	// Connection timeout to the database
	Timeout int `json:"timeout,omitempty"`

	// Use TLS to connect to the store
	UseTls bool `json:"useTls,omitempty"`

	// Allow invalid TLS certificates when connecting to the store
	AllowInvalidCerts bool `json:"allowInvalidCerts,omitempty"`

	// Maximum size of a packet in bytes
	// +kubebuilder:validation:Minimum=1024
	// +kubebuilder:validation:Maximum=1073741824
	MaxAllowedPacket int `json:"maxAllowedPacket,omitempty"`

	// Maximum number of connections to the store
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=8192
	PoolMaxConnections int `json:"poolMaxConnections,omitempty"`

	// Minimum number of connections to the store
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=8192
	PoolMinConnections int `json:"poolMinConnections,omitempty"`

	// Hostname of the database server
	// +required
	Host string `json:"host"`

	// Port of the database server
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port,omitempty"`

	// Name of the database
	Database string `json:"database,omitempty"`

	// Username to connect to the store
	AuthUsername string `json:"authUsername,omitempty"`

	// Password to connect to the store
	// +required
	AuthSecret string `json:"authSecret"`
}

func assertValue[T any](config *T, storeType string) (any, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration for datastore of type %q is missing", storeType)
	}
	return config, nil
}

func (dataStore StalwartDataStore) values() (any, error) {
	switch dataStore.Type {
	case "RocksDb":
		return assertValue(dataStore.RocksDb, dataStore.Type)
	case "Sqlite":
		return assertValue(dataStore.Sqlite, dataStore.Type)
	case "FoundationDb":
		return assertValue(dataStore.FoundationDb, dataStore.Type)
	case "PostgreSql":
		return assertValue(dataStore.PostgreSql, dataStore.Type)
	case "MySql":
		return assertValue(dataStore.MySql, dataStore.Type)
	default:
		return nil, fmt.Errorf("unsupported datastore type %q", dataStore.Type)
	}
}

func (dataStore StalwartDataStore) ToStalwartConfig() ([]byte, error) {
	values, err := dataStore.values()
	if err != nil {
		return nil, err
	}

	config, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("failed converting datastore config: %w", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(config, &fields); err != nil {
		return nil, fmt.Errorf("failed converting datastore config: %w", err)
	}

	fields["@type"] = json.RawMessage(strconv.Quote(dataStore.Type))
	return json.Marshal(fields)
}
