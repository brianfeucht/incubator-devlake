/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/linear/models"
)

// TestConnectionResult is the response body for connection test endpoints.
type TestConnectionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// TestConnection validates a new Linear connection (before saving).
func TestConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	connection := &models.LinearConnection{}
	// Use nil validator — Name is not required for a connection test
	if err := helper.Decode(input.Body, connection, nil); err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}
	connection.Normalize()
	if err := validateConnection(connection); err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}
	result, err := testLinearConnection(connection)
	if err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}
	return &plugin.ApiResourceOutput{Body: result, Status: http.StatusOK}, nil
}

// TestExistingConnection validates a saved Linear connection by ID.
func TestExistingConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	connection := &models.LinearConnection{}
	if err := connectionHelper.First(connection, input.Params); err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, errors.BadInput.Wrap(err, "find connection from db"))
	}
	if err := helper.DecodeMapStruct(input.Body, connection, false); err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}
	connection.Normalize()
	if err := validateConnection(connection); err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}
	result, err := testLinearConnection(connection)
	if err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}
	return &plugin.ApiResourceOutput{Body: result, Status: http.StatusOK}, nil
}

// testLinearConnection sends a viewer query to verify the API key.
func testLinearConnection(connection *models.LinearConnection) (*TestConnectionResult, errors.Error) {
	apiClient, err := helper.NewApiClientFromConnection(context.Background(), basicRes, connection)
	if err != nil {
		return nil, err
	}
	apiClient.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	})

	res, err := apiClient.Post("graphql", nil, strings.NewReader(`{"query":"{ viewer { id name email } }"}`), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return nil, errors.Unauthorized.New("invalid Linear API key")
	}
	if res.StatusCode >= 400 {
		return nil, errors.HttpStatus(res.StatusCode).New("Linear API returned an error")
	}

	return &TestConnectionResult{Success: true, Message: "Connection successful"}, nil
}
