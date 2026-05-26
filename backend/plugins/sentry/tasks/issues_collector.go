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
	"time"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
)

const RAW_ISSUE_TABLE = "sentry_api_issues"

var CollectIssuesMeta = plugin.SubTaskMeta{
	Name:             "CollectIssues",
	EntryPoint:       CollectIssues,
	EnabledByDefault: true,
	Description:      "Collect Sentry issues (groups) for a project — supports incremental sync",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{},
	ProductTables:    []string{RAW_ISSUE_TABLE},
}

func CollectIssues(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*SentryTaskData)

	apiCollector, err := helper.NewStatefulApiCollector(helper.RawDataSubTaskArgs{
		Ctx: taskCtx,
		Params: SentryApiParams{
			ConnectionId: data.Options.ConnectionId,
			ScopeId:      data.Options.ScopeId,
		},
		Table: RAW_ISSUE_TABLE,
	})
	if err != nil {
		return err
	}

	since := apiCollector.GetSince()

	err = apiCollector.InitCollector(helper.ApiCollectorArgs{
		ApiClient: data.ApiClient,
		PageSize:  100,
		UrlTemplate: fmt.Sprintf(
			"api/0/projects/%s/%s/issues/",
			data.Options.OrgSlug,
			data.Options.ProjectSlug,
		),
		Query: func(reqData *helper.RequestData) (url.Values, errors.Error) {
			query := url.Values{}
			query.Set("query", "")
			query.Set("limit", fmt.Sprintf("%d", reqData.Pager.Size))
			if since != nil {
				// Sentry accepts ISO 8601 for `start` param
				query.Set("start", since.UTC().Format(time.RFC3339))
			}
			// Use cursor-based pagination from the Link header
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
