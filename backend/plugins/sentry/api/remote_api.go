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
	"fmt"
	"net/url"
	"strings"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	dsmodels "github.com/apache/incubator-devlake/helpers/pluginhelper/api/models"
	"github.com/apache/incubator-devlake/plugins/sentry/models"
)

// SentryRemotePagination carries the current page cursor for scope listing.
type SentryRemotePagination struct {
	Cursor string `json:"cursor"`
}

// sentryProjectResponse is a minimal projection of the Sentry project JSON.
type sentryProjectResponse struct {
	ID       string `json:"id"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
}

func listSentryRemoteScopes(
	connection *models.SentryConnection,
	apiClient plugin.ApiClient,
	groupId string,
	pagination SentryRemotePagination,
) (
	children []dsmodels.DsRemoteApiScopeListEntry[models.SentryProject],
	nextPage *SentryRemotePagination,
	err errors.Error,
) {
	if connection == nil {
		return nil, nil, errors.BadInput.New("connection is required")
	}

	// groupId is the org slug selected by the user in the UI (passed via query param).
	orgSlug := strings.TrimSpace(groupId)
	if orgSlug == "" {
		// No org selected yet — return the orgs list as group entries.
		return listOrgsAsGroups(apiClient, connection)
	}

	// Org selected — return its projects.
	query := url.Values{}
	query.Set("per_page", "100")
	if pagination.Cursor != "" {
		query.Set("cursor", pagination.Cursor)
	}

	res, apiErr := apiClient.Get(fmt.Sprintf("api/0/organizations/%s/projects/", orgSlug), query, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}

	var projects []sentryProjectResponse
	if apiErr = api.UnmarshalResponse(res, &projects); apiErr != nil {
		res.Body.Close()
		return nil, nil, apiErr
	}
	res.Body.Close()

	for _, p := range projects {
		scopeId := orgSlug + "/" + p.Slug
		children = append(children, dsmodels.DsRemoteApiScopeListEntry[models.SentryProject]{
			Type:     api.RAS_ENTRY_TYPE_SCOPE,
			Id:       scopeId,
			Name:     p.Name,
			FullName: scopeId,
			Data: &models.SentryProject{
				Id:       scopeId,
				OrgSlug:  orgSlug,
				Slug:     p.Slug,
				Name:     p.Name,
				Platform: p.Platform,
			},
		})
	}

	// Parse next cursor from Link header.
	if next := parseNextCursorFromLink(res.Header.Get("Link")); next != "" {
		nextPage = &SentryRemotePagination{Cursor: next}
	}

	return children, nextPage, nil
}

// listOrgsAsGroups fetches the user's organizations and returns them as GROUP entries
// so the UI can drill into a specific org to see its projects.
func listOrgsAsGroups(
	apiClient plugin.ApiClient,
	_ *models.SentryConnection,
) (
	children []dsmodels.DsRemoteApiScopeListEntry[models.SentryProject],
	nextPage *SentryRemotePagination,
	err errors.Error,
) {
	res, apiErr := apiClient.Get("api/0/organizations/", url.Values{"per_page": []string{"100"}}, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}

	var orgs []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if apiErr = api.UnmarshalResponse(res, &orgs); apiErr != nil {
		res.Body.Close()
		return nil, nil, apiErr
	}
	res.Body.Close()

	for _, o := range orgs {
		children = append(children, dsmodels.DsRemoteApiScopeListEntry[models.SentryProject]{
			Type:     api.RAS_ENTRY_TYPE_GROUP,
			Id:       o.Slug,
			Name:     o.Name,
			FullName: o.Slug,
		})
	}
	return children, nil, nil
}

func searchSentryRemoteScopes(
	apiClient plugin.ApiClient,
	params *dsmodels.DsRemoteApiScopeSearchParams,
) (
	children []dsmodels.DsRemoteApiScopeListEntry[models.SentryProject],
	err errors.Error,
) {
	if params == nil || strings.TrimSpace(params.Search) == "" {
		return nil, nil
	}
	// Sentry does not expose a free-text project search across all orgs.
	// We list all orgs, then list all projects per org and filter locally.
	res, apiErr := apiClient.Get("api/0/organizations/", url.Values{"per_page": []string{"100"}}, nil)
	if apiErr != nil {
		return nil, apiErr
	}

	var orgs []struct {
		Slug string `json:"slug"`
	}
	if apiErr = api.UnmarshalResponse(res, &orgs); apiErr != nil {
		res.Body.Close()
		return nil, apiErr
	}
	res.Body.Close()

	queryLower := strings.ToLower(strings.TrimSpace(params.Search))

	for _, org := range orgs {
		projRes, apiErr := apiClient.Get(
			fmt.Sprintf("api/0/organizations/%s/projects/", org.Slug),
			url.Values{"per_page": []string{"100"}},
			nil,
		)
		if apiErr != nil {
			continue
		}

		var projects []sentryProjectResponse
		if apiErr = api.UnmarshalResponse(projRes, &projects); apiErr != nil {
			projRes.Body.Close()
			continue
		}
		projRes.Body.Close()

		for _, p := range projects {
			if !strings.Contains(strings.ToLower(p.Name), queryLower) &&
				!strings.Contains(strings.ToLower(p.Slug), queryLower) {
				continue
			}
			scopeId := org.Slug + "/" + p.Slug
			children = append(children, dsmodels.DsRemoteApiScopeListEntry[models.SentryProject]{
				Type:     api.RAS_ENTRY_TYPE_SCOPE,
				Id:       scopeId,
				Name:     p.Name,
				FullName: scopeId,
				Data: &models.SentryProject{
					Id:       scopeId,
					OrgSlug:  org.Slug,
					Slug:     p.Slug,
					Name:     p.Name,
					Platform: p.Platform,
				},
			})
		}
	}

	return children, nil
}

// RemoteScopes lists available Sentry projects for this connection.
// @Summary list available Sentry projects
// @Description list available Sentry projects (scopes) for this connection
// @Tags plugins/sentry
// @Accept application/json
// @Param connectionId path int false "connection ID"
// @Param groupId query string false "org slug"
// @Param pageToken query string false "page token (cursor)"
// @Success 200 {object} dsmodels.DsRemoteApiScopeList[models.SentryProject]
// @Failure 400 {object} shared.ApiBody "Bad Request"
// @Failure 500 {object} shared.ApiBody "Internal Error"
// @Router /plugins/sentry/connections/{connectionId}/remote-scopes [GET]
func RemoteScopes(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return raScopeList.Get(input)
}

// SearchRemoteScopes searches for Sentry projects by name.
// @Summary search Sentry projects
// @Description search available Sentry projects by name or slug
// @Tags plugins/sentry
// @Accept application/json
// @Param connectionId path int false "connection ID"
// @Param search query string false "search term"
// @Param page query int false "page number"
// @Param pageSize query int false "page size"
// @Success 200 {object} dsmodels.DsRemoteApiScopeList[models.SentryProject]
// @Failure 400 {object} shared.ApiBody "Bad Request"
// @Failure 500 {object} shared.ApiBody "Internal Error"
// @Router /plugins/sentry/connections/{connectionId}/search-remote-scopes [GET]
func SearchRemoteScopes(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return raScopeSearch.Get(input)
}

// parseNextCursorFromLink extracts the next cursor value from a Sentry Link header.
func parseNextCursorFromLink(link string) string {
	if link == "" {
		return ""
	}
	for _, part := range strings.Split(link, ", ") {
		if !strings.Contains(part, `rel="next"`) || !strings.Contains(part, `results="true"`) {
			continue
		}
		// Extract cursor="..." value
		const prefix = `cursor="`
		idx := strings.Index(part, prefix)
		if idx < 0 {
			break
		}
		rest := part[idx+len(prefix):]
		end := strings.Index(rest, `"`)
		if end < 0 {
			break
		}
		return rest[:end]
	}
	return ""
}

// Proxy forwards arbitrary Sentry API requests (used by the UI for dynamic data).
func Proxy(input *plugin.ApiResourceInput) (*plugin.ApiResourceOutput, errors.Error) {
	return raProxy.Proxy(input)
}
