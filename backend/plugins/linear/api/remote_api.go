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
	"strings"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	pluginApi "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	dsmodels "github.com/apache/incubator-devlake/helpers/pluginhelper/api/models"
	"github.com/apache/incubator-devlake/plugins/linear/models"
)

// LinearRemotePagination carries the cursor for scope listing.
type LinearRemotePagination struct {
	Cursor string `json:"cursor"`
}

// linearTeamResponse is a minimal projection of a Linear team.
type linearTeamResponse struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
}

func listLinearRemoteScopes(
	connection *models.LinearConnection,
	apiClient plugin.ApiClient,
	_ string,
	pagination LinearRemotePagination,
) (
	children []dsmodels.DsRemoteApiScopeListEntry[models.LinearTeam],
	nextPage *LinearRemotePagination,
	err errors.Error,
) {
	afterClause := ""
	if pagination.Cursor != "" {
		afterClause = `, after: "` + pagination.Cursor + `"`
	}
	gqlQuery := `{ teams(first: 50` + afterClause + `) { nodes { id name key description } pageInfo { hasNextPage endCursor } } }`

	res, apiErr := apiClient.Post("graphql", nil, map[string]interface{}{"query": gqlQuery}, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}
	defer res.Body.Close()

	type pageInfo struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor"`
	}
	type teamsConn struct {
		Nodes    []linearTeamResponse `json:"nodes"`
		PageInfo pageInfo             `json:"pageInfo"`
	}
	type dataWrapper struct {
		Teams teamsConn `json:"teams"`
	}
	type gqlResp struct {
		Data dataWrapper `json:"data"`
	}

	var resp gqlResp
	if apiErr = pluginApi.UnmarshalResponse(res, &resp); apiErr != nil {
		return nil, nil, apiErr
	}

	for _, t := range resp.Data.Teams.Nodes {
		children = append(children, dsmodels.DsRemoteApiScopeListEntry[models.LinearTeam]{
			Type:     pluginApi.RAS_ENTRY_TYPE_SCOPE,
			Id:       t.Id,
			Name:     t.Name,
			FullName: t.Name,
			Data: &models.LinearTeam{
				Id:          t.Id,
				Name:        t.Name,
				Key:         t.Key,
				Description: t.Description,
			},
		})
	}

	if resp.Data.Teams.PageInfo.HasNextPage {
		nextPage = &LinearRemotePagination{Cursor: resp.Data.Teams.PageInfo.EndCursor}
	}

	return children, nextPage, nil
}

func searchLinearRemoteScopes(
	apiClient plugin.ApiClient,
	params *dsmodels.DsRemoteApiScopeSearchParams,
) (
	children []dsmodels.DsRemoteApiScopeListEntry[models.LinearTeam],
	err errors.Error,
) {
	if params == nil || strings.TrimSpace(params.Search) == "" {
		return nil, nil
	}

	// Fetch all teams (up to 250) and filter locally
	res, apiErr := apiClient.Post("graphql", nil, map[string]interface{}{"query": "{ teams(first: 250) { nodes { id name key description } } }"}, nil)
	if apiErr != nil {
		return nil, apiErr
	}
	defer res.Body.Close()

	type teamsConn struct {
		Nodes []linearTeamResponse `json:"nodes"`
	}
	type dataWrapper struct {
		Teams teamsConn `json:"teams"`
	}
	type gqlResp struct {
		Data dataWrapper `json:"data"`
	}
	var resp gqlResp
	if apiErr = pluginApi.UnmarshalResponse(res, &resp); apiErr != nil {
		return nil, apiErr
	}

	queryLower := strings.ToLower(strings.TrimSpace(params.Search))
	for _, t := range resp.Data.Teams.Nodes {
		if !strings.Contains(strings.ToLower(t.Name), queryLower) &&
			!strings.Contains(strings.ToLower(t.Key), queryLower) {
			continue
		}
		children = append(children, dsmodels.DsRemoteApiScopeListEntry[models.LinearTeam]{
			Type:     pluginApi.RAS_ENTRY_TYPE_SCOPE,
			Id:       t.Id,
			Name:     t.Name,
			FullName: t.Name,
			Data: &models.LinearTeam{
				Id:          t.Id,
				Name:        t.Name,
				Key:         t.Key,
				Description: t.Description,
			},
		})
	}
	return children, nil
}

// RemoteScopes lists available Linear teams for this connection.
func RemoteScopes(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return raScopeList.Get(input)
}

// SearchRemoteScopes searches for Linear teams by name.
func SearchRemoteScopes(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return raScopeSearch.Get(input)
}

// Proxy forwards arbitrary Linear API requests.
func Proxy(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return raProxy.Proxy(input)
}
