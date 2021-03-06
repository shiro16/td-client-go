//
// Treasure Data API client for Go
//
// Copyright (C) 2014 Treasure Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package td_client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"time"
)

// ListDataBasesResultElement represents an item of the result of
// ListDatabases API call
type ListDataBasesResultElement struct {
	Name         string
	Organization string
	Count        int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Permission   string
}

// ListDataBasesResult is a collection of ListDataBasesResultElement
type ListDataBasesResult []ListDataBasesResultElement

var listDatabasesSchema = map[string]interface{}{
	"databases": []map[string]interface{}{
		map[string]interface{}{
			"name":         "",
			"organization": Optional{"", ""},
			"count":        0,
			"created_at":   time.Time{},
			"updated_at":   time.Time{},
			"permission":   "",
		},
	},
}

// ListTablesResultElement represents an item of the result of ListTables API
// call
type ListTablesResultElement struct {
	Id                   int
	Name                 string
	Type                 string
	Count                int
	CreatedAt            time.Time
	UpdatedAt            time.Time
	LastImport           time.Time
	LastLogTimestamp     time.Time
	EstimatedStorageSize int
	Schema               []interface{}
	ExpireDays           int
	PrimaryKey           string
	PrimaryKeyType       string
}

// ListTablesResult is a collection of ListTablesResultElement
type ListTablesResult []ListTablesResultElement

var listTablesSchema = map[string]interface{}{
	"database": "",
	"tables": []map[string]interface{}{
		map[string]interface{}{
			"id":                     0,
			"name":                   "",
			"type":                   Optional{"", "?"},
			"count":                  Optional{0, 0},
			"created_at":             time.Time{},
			"updated_at":             time.Time{},
			"counter_updated_at":     Optional{time.Time{}, time.Time{}},
			"last_log_timestamp":     Optional{time.Time{}, time.Time{}},
			"estimated_storage_size": 0,
			"schema":                 Optional{EmbeddedJSON([]interface{}{}), nil},
			"expire_days":            Optional{0, 0},
			"primary_key":            Optional{"", ""},
			"primary_key_type":       Optional{"", ""},
		},
	},
}

var deleteTableSchema = map[string]interface{}{
	"table":    "",
	"database": "",
	"type":     Optional{"", "?"},
}

func (client *TDClient) ListDatabases() (*ListDataBasesResult, error) {
	resp, err := client.get("/v3/database/list", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, client.buildError(resp, -1, "List databases failed", nil)
	}
	js, err := client.checkedJson(resp, listDatabasesSchema)
	if err != nil {
		return nil, err
	}
	databases := js["databases"].([]map[string]interface{})
	retval := make(ListDataBasesResult, len(databases))
	for i, v := range databases {
		retval[i] = ListDataBasesResultElement{
			Name:         v["name"].(string),
			Organization: v["organization"].(string),
			Count:        v["count"].(int),
			CreatedAt:    v["created_at"].(time.Time),
			UpdatedAt:    v["updated_at"].(time.Time),
			Permission:   v["permission"].(string),
		}
	}
	return &retval, nil
}

func (client *TDClient) DeleteDatabase(db string) error {
	resp, err := client.post(fmt.Sprintf("/v3/database/delete/%s", url.QueryEscape(db)), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return client.buildError(resp, -1, "Delete database failed", nil)
	}
	return nil
}

func (client *TDClient) CreateDatabase(db string, options map[string]string) error {
	resp, err := client.post(fmt.Sprintf("/v3/database/create/%s", url.QueryEscape(db)), dictToValues(options))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return client.buildError(resp, -1, "Create database failed", nil)
	}
	return nil
}

func (client *TDClient) ListTables(db string) (*ListTablesResult, error) {
	resp, err := client.get(fmt.Sprintf("/v3/table/list/%s", url.QueryEscape(db)), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, client.buildError(resp, -1, "List tables failed", nil)
	}
	js, err := client.checkedJson(resp, listTablesSchema)
	if err != nil {
		return nil, err
	}
	tables := js["tables"].([]map[string]interface{})
	retval := make(ListTablesResult, len(tables))
	for i, v := range tables {
		retval[i] = ListTablesResultElement{
			Id:                   v["id"].(int),
			Name:                 v["name"].(string),
			Type:                 v["type"].(string),
			Count:                v["count"].(int),
			CreatedAt:            v["created_at"].(time.Time),
			UpdatedAt:            v["updated_at"].(time.Time),
			LastImport:           v["counter_updated_at"].(time.Time),
			LastLogTimestamp:     v["last_log_timestamp"].(time.Time),
			EstimatedStorageSize: v["estimated_storage_size"].(int),
			Schema:               v["schema"].([]interface{}),
			ExpireDays:           v["expire_days"].(int),
			PrimaryKey:           v["primary_key"].(string),
			PrimaryKeyType:       v["primary_key_type"].(string),
		}
	}
	return &retval, nil
}

func (client *TDClient) createTable(db string, table string, type_ string, params map[string]string) error {
	resp, err := client.post(fmt.Sprintf("/v3/table/create/%s/%s/%s", url.QueryEscape(db), url.QueryEscape(table), url.QueryEscape(type_)), dictToValues(params))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return client.buildError(resp, -1, fmt.Sprintf("Create %s table failed", type_), nil)
	}
	return nil
}

func (client *TDClient) CreateItemTable(db string, table string, primaryKey string, primaryKeyType string) error {
	return client.createTable(
		db, table, "item",
		map[string]string{
			"primary_key":      primaryKey,
			"primary_key_type": primaryKeyType,
		},
	)
}

func (client *TDClient) CreateLogTable(db string, table string) error {
	return client.createTable(db, table, "log", nil)
}

func (client *TDClient) SwapTable(db string, table1 string, table2 string) error {
	resp, err := client.post(fmt.Sprintf("/v3/table/swap/%s/%s/%s", url.QueryEscape(db), url.QueryEscape(table1), url.QueryEscape(table2)), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return client.buildError(resp, -1, "Swap tables failed", nil)
	}
	return nil
}

func (client *TDClient) UpdateSchema(db string, table string, schema []interface{}) error {
	jsStr, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	resp, err := client.post(fmt.Sprintf("/v3/table/update-schema/%s/%s", url.QueryEscape(db), url.QueryEscape(table)), url.Values{"schema": {string(jsStr)}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return client.buildError(resp, -1, "Update schema failed", nil)
	}
	return nil
}

func (client *TDClient) UpdateExpire(db string, table string, expireDays int) error {
	resp, err := client.post(fmt.Sprintf("/v3/table/update/%s/%s", url.QueryEscape(db), url.QueryEscape(table)), url.Values{"expire_days": {strconv.Itoa(expireDays)}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return client.buildError(resp, -1, "Update expire failed", nil)
	}
	return nil
}

func (client *TDClient) DeleteTable(db string, table string) (string, error) {
	resp, err := client.post(fmt.Sprintf("/v3/table/delete/%s/%s", url.QueryEscape(db), url.QueryEscape(table)), nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", client.buildError(resp, -1, "Delete table failed", nil)
	}
	js, err := client.checkedJson(resp, deleteTableSchema)
	if err != nil {
		return "", err
	}
	return js["type"].(string), err
}

func (client *TDClient) Tail(db string, table string, count int, to time.Time, from time.Time, reader func(interface{}) error) error {
	params := url.Values{}
	if count > 0 {
		params.Set("count", strconv.Itoa(count))
	}
	if !to.IsZero() {
		params.Set("to", to.UTC().Format(TDAPIDateTime))
	}
	if !from.IsZero() {
		params.Set("from", from.UTC().Format(TDAPIDateTime))
	}
	resp, err := client.post(fmt.Sprintf("/v3/table/tail/%s/%s", url.QueryEscape(db), url.QueryEscape(table)), params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return client.buildError(resp, -1, "Tail failed", nil)
	}
	dec := client.getMessagePackDecoder(resp.Body)
	for {
		v := (interface{})(nil)
		err := dec.Decode(&v)
		if err != nil {
			if err == io.EOF {
				break
			}
			return client.buildError(resp, -1, "Invalid MessagePack stream", nil)
		}
		err = reader(v)
		if err != nil {
			return client.buildError(resp, -1, "Reader returned error status", err)
		}
	}
	return nil
}
