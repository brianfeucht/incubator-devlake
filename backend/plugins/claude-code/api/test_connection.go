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

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/claude-code/models"
)

// TestConnectionResult is the response body for connection test endpoints.
type TestConnectionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// TestConnection validates a new Claude connection (before saving).
func TestConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	connection := &models.ClaudeConnection{}
	// Use nil validator — Name is not required for a connection test
	if err := helper.Decode(input.Body, connection, nil); err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}
	connection.Normalize()
	if err := validateConnection(connection); err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}
	result, err := testClaudeConnection(connection)
	if err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}
	return &plugin.ApiResourceOutput{Body: result, Status: http.StatusOK}, nil
}

// TestExistingConnection validates a saved Claude connection by ID.
func TestExistingConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	connection := &models.ClaudeConnection{}
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
	result, err := testClaudeConnection(connection)
	if err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}
	return &plugin.ApiResourceOutput{Body: result, Status: http.StatusOK}, nil
}

// testClaudeConnection calls the Anthropic API to verify the key is valid.
// First checks connectivity with /v1/models, then verifies Admin API access via /v1/organizations/workspaces.
func testClaudeConnection(connection *models.ClaudeConnection) (*TestConnectionResult, errors.Error) {
	apiClient, err := helper.NewApiClientFromConnection(context.Background(), basicRes, connection)
	if err != nil {
		return nil, err
	}
	apiClient.SetHeaders(map[string]string{"Accept": "application/json"})

	// Step 1: validate key is valid via /v1/models
	modelsRes, err := apiClient.Get("v1/models", nil, nil)
	if err != nil {
		return nil, err
	}
	defer modelsRes.Body.Close()
	if modelsRes.StatusCode == http.StatusUnauthorized {
		return nil, errors.Unauthorized.New("invalid API key — please provide a valid Anthropic API key")
	}
	if modelsRes.StatusCode >= 400 {
		return nil, errors.HttpStatus(modelsRes.StatusCode).New("Anthropic API returned an error on /v1/models")
	}

	// Step 2: verify Admin API access required for cost/usage collection
	adminRes, adminErr := apiClient.Get("v1/organizations/workspaces", nil, nil)
	if adminErr != nil {
		return nil, adminErr
	}
	defer adminRes.Body.Close()
	if adminRes.StatusCode == http.StatusUnauthorized || adminRes.StatusCode == http.StatusForbidden {
		return nil, errors.Unauthorized.New(
			"API key is valid but lacks Admin API access. " +
				"Cost and usage collection requires an Admin API key (sk-ant-admin-...). " +
				"Create one at https://console.anthropic.com under Settings > API Keys.",
		)
	}
	if adminRes.StatusCode >= 400 {
		return nil, errors.HttpStatus(adminRes.StatusCode).New("Anthropic Admin API returned an error on /v1/organizations/workspaces")
	}

	return &TestConnectionResult{Success: true, Message: "Connection successful — Admin API access confirmed"}, nil
}
