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

package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
)

const RAW_PROJECT_TABLE = "sentry_api_projects"

var CollectProjectsMeta = plugin.SubTaskMeta{
	Name:             "CollectProjects",
	EntryPoint:       CollectProjects,
	EnabledByDefault: true,
	Description:      "Collect Sentry projects for an organization",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{},
	ProductTables:    []string{RAW_PROJECT_TABLE},
}

func CollectProjects(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*SentryTaskData)

	apiCollector, err := helper.NewStatefulApiCollector(helper.RawDataSubTaskArgs{
		Ctx: taskCtx,
		Params: SentryApiParams{
			ConnectionId: data.Options.ConnectionId,
			ScopeId:      data.Options.ScopeId,
		},
		Table: RAW_PROJECT_TABLE,
	})
	if err != nil {
		return err
	}

	err = apiCollector.InitCollector(helper.ApiCollectorArgs{
		ApiClient: data.ApiClient,
		PageSize:  100,
		UrlTemplate: fmt.Sprintf("api/0/organizations/%s/projects/", data.Options.OrgSlug),
		Query: func(reqData *helper.RequestData) (url.Values, errors.Error) {
			query := url.Values{}
			query.Set("per_page", fmt.Sprintf("%d", reqData.Pager.Size))
			if reqData.CustomData != nil {
				if cursor, ok := reqData.CustomData.(string); ok && cursor != "" {
					query.Set("cursor", cursor)
				}
			}
			return query, nil
		},
		GetNextPageCustomData: parseSentryLinkCursor,
		ResponseParser: func(res *http.Response) ([]json.RawMessage, errors.Error) {
			var items []json.RawMessage
			if err := helper.UnmarshalResponse(res, &items); err != nil {
				return nil, err
			}
			return items, nil
		},
	})
	if err != nil {
		return err
	}

	return apiCollector.Execute()
}

// SentryApiParams uniquely identifies the raw data source for state management.
type SentryApiParams struct {
	ConnectionId uint64 `json:"connectionId"`
	ScopeId      string `json:"scopeId"`
}
