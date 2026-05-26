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
	"github.com/apache/incubator-devlake/plugins/sentry/models"
)

// TestConnectionResult is the response body for connection test endpoints.
type TestConnectionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// TestConnection validates a new Sentry connection (before saving).
func TestConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	connection := &models.SentryConnection{}
	if err := helper.Decode(input.Body, connection, vld); err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}

	connection.Normalize()
	if err := validateConnection(connection); err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}

	result, err := testSentryConnection(basicRes, connection)
	if err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}
	return &plugin.ApiResourceOutput{Body: result, Status: http.StatusOK}, nil
}

// TestExistingConnection validates a saved Sentry connection by ID.
func TestExistingConnection(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	connection := &models.SentryConnection{}
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

	result, err := testSentryConnection(basicRes, connection)
	if err != nil {
		return nil, plugin.WrapTestConnectionErrResp(basicRes, err)
	}
	return &plugin.ApiResourceOutput{Body: result, Status: http.StatusOK}, nil
}

// testSentryConnection hits the Sentry API to verify the token is valid.
func testSentryConnection(_ interface{}, connection *models.SentryConnection) (*TestConnectionResult, errors.Error) {
	apiClient, err := helper.NewApiClientFromConnection(context.Background(), nil, connection)
	if err != nil {
		return nil, err
	}
	apiClient.SetHeaders(map[string]string{"Accept": "application/json"})

	res, err := apiClient.Get("api/0/organizations/", nil, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
		return nil, errors.Unauthorized.New("invalid or insufficient Sentry token")
	}
	if res.StatusCode >= 400 {
		return nil, errors.HttpStatus(res.StatusCode).New("Sentry API returned an error")
	}

	return &TestConnectionResult{Success: true, Message: "Connection successful"}, nil
}
