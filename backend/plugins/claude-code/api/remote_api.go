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
	"net/url"
	"strings"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	dsmodels "github.com/apache/incubator-devlake/helpers/pluginhelper/api/models"
	"github.com/apache/incubator-devlake/plugins/claude-code/models"
)

// ClaudeRemotePagination carries the next_page cursor for workspace listing.
type ClaudeRemotePagination struct {
	NextPage string `json:"nextPage"`
}

// claudeWorkspaceResponse is a minimal projection of the Anthropic workspace JSON.
type claudeWorkspaceResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// claudeWorkspaceListResponse is the paginated list response.
type claudeWorkspaceListResponse struct {
	Data     []claudeWorkspaceResponse `json:"data"`
	HasMore  bool                      `json:"has_more"`
	NextPage string                    `json:"next_page"`
}

func listClaudeRemoteScopes(
	_ *models.ClaudeConnection,
	apiClient plugin.ApiClient,
	_ string,
	pagination ClaudeRemotePagination,
) (
	children []dsmodels.DsRemoteApiScopeListEntry[models.ClaudeWorkspace],
	nextPage *ClaudeRemotePagination,
	err errors.Error,
) {
	query := url.Values{}
	query.Set("limit", "100")
	if pagination.NextPage != "" {
		query.Set("after_id", pagination.NextPage)
	}

	res, apiErr := apiClient.Get("v1/organizations/workspaces", query, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}

	var listResp claudeWorkspaceListResponse
	if apiErr = api.UnmarshalResponse(res, &listResp); apiErr != nil {
		res.Body.Close()
		return nil, nil, apiErr
	}
	res.Body.Close()

	for _, w := range listResp.Data {
		displayName := w.DisplayName
		if displayName == "" {
			displayName = w.Name
		}
		children = append(children, dsmodels.DsRemoteApiScopeListEntry[models.ClaudeWorkspace]{
			Type:     api.RAS_ENTRY_TYPE_SCOPE,
			Id:       w.ID,
			Name:     displayName,
			FullName: w.ID,
			Data: &models.ClaudeWorkspace{
				Id:          w.ID,
				Name:        w.Name,
				DisplayName: displayName,
			},
		})
	}

	if listResp.HasMore && listResp.NextPage != "" {
		nextPage = &ClaudeRemotePagination{NextPage: listResp.NextPage}
	}

	return children, nextPage, nil
}

func searchClaudeRemoteScopes(
	apiClient plugin.ApiClient,
	params *dsmodels.DsRemoteApiScopeSearchParams,
) (
	children []dsmodels.DsRemoteApiScopeListEntry[models.ClaudeWorkspace],
	err errors.Error,
) {
	if params == nil || strings.TrimSpace(params.Search) == "" {
		return nil, nil
	}

	// Fetch all workspaces and filter locally (Anthropic API has no free-text search).
	query := url.Values{"limit": []string{"100"}}
	res, apiErr := apiClient.Get("v1/organizations/workspaces", query, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	var listResp claudeWorkspaceListResponse
	if apiErr = api.UnmarshalResponse(res, &listResp); apiErr != nil {
		res.Body.Close()
		return nil, apiErr
	}
	res.Body.Close()

	queryLower := strings.ToLower(strings.TrimSpace(params.Search))
	for _, w := range listResp.Data {
		displayName := w.DisplayName
		if displayName == "" {
			displayName = w.Name
		}
		if !strings.Contains(strings.ToLower(displayName), queryLower) &&
			!strings.Contains(strings.ToLower(w.ID), queryLower) {
			continue
		}
		children = append(children, dsmodels.DsRemoteApiScopeListEntry[models.ClaudeWorkspace]{
			Type:     api.RAS_ENTRY_TYPE_SCOPE,
			Id:       w.ID,
			Name:     displayName,
			FullName: w.ID,
			Data: &models.ClaudeWorkspace{
				Id:          w.ID,
				Name:        w.Name,
				DisplayName: displayName,
			},
		})
	}

	return children, nil
}

// RemoteScopes lists available Anthropic workspaces for this connection.
func RemoteScopes(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return raScopeList.Get(input)
}

// SearchRemoteScopes searches for Anthropic workspaces by name.
func SearchRemoteScopes(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return raScopeSearch.Get(input)
}

// Proxy forwards arbitrary Anthropic API requests.
func Proxy(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return raProxy.Proxy(input)
}
