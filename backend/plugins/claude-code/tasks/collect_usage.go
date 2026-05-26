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
	"net/http"
	"net/url"
	"time"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
)

const RAW_USAGE_TABLE = "claude_api_usage_report"

var CollectUsageReportMeta = plugin.SubTaskMeta{
	Name:             "CollectUsageReport",
	EntryPoint:       CollectUsageReport,
	EnabledByDefault: true,
	Description:      "Collect Anthropic token usage report data (daily buckets per model/workspace/user)",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{},
	ProductTables:    []string{RAW_USAGE_TABLE},
}

// claudeUsageReportEnvelope is the top-level usage report response.
type claudeUsageReportEnvelope struct {
	Data     []claudeUsageBucket `json:"data"`
	HasMore  bool                `json:"has_more"`
	NextPage string              `json:"next_page"`
}

type claudeUsageBucket struct {
	StartingAt string            `json:"starting_at"`
	EndingAt   string            `json:"ending_at"`
	Results    []json.RawMessage `json:"results"`
}

func CollectUsageReport(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*ClaudeTaskData)

	apiCollector, err := helper.NewStatefulApiCollector(helper.RawDataSubTaskArgs{
		Ctx: taskCtx,
		Params: claudeApiParams{
			ConnectionId: data.Options.ConnectionId,
			ScopeId:      data.Options.ScopeId,
		},
		Table: RAW_USAGE_TABLE,
	})
	if err != nil {
		return err
	}

	since := apiCollector.GetSince()
	daysToSync := data.Options.DaysToSync
	if daysToSync <= 0 {
		daysToSync = 30
	}

	startDate := time.Now().UTC().AddDate(0, 0, -daysToSync).Format("2006-01-02")
	if since != nil {
		startDate = since.UTC().Format("2006-01-02")
	}
	endDate := time.Now().UTC().Format("2006-01-02")

	var nextPageCursor string

	err = apiCollector.InitCollector(helper.ApiCollectorArgs{
		ApiClient:   data.ApiClient,
		PageSize:    100,
		UrlTemplate: "v1/organizations/usage_report/messages",
		Query: func(reqData *helper.RequestData) (url.Values, errors.Error) {
			query := url.Values{}
			query.Set("start_date", startDate)
			query.Set("end_date", endDate)
			query.Set("time_period", "day")
			if data.Options.ScopeId != "" {
				query.Set("workspace_ids[]", data.Options.ScopeId)
			}
			if reqData.CustomData != nil {
				if cursor, ok := reqData.CustomData.(string); ok && cursor != "" {
					query.Set("next_page", cursor)
				}
			}
			return query, nil
		},
		GetNextPageCustomData: func(_ *helper.RequestData, _ *http.Response) (interface{}, errors.Error) {
			if nextPageCursor == "" {
				return nil, nil
			}
			c := nextPageCursor
			nextPageCursor = ""
			return c, nil
		},
		ResponseParser: func(res *http.Response) ([]json.RawMessage, errors.Error) {
			var envelope claudeUsageReportEnvelope
			if parseErr := helper.UnmarshalResponse(res, &envelope); parseErr != nil {
				return nil, parseErr
			}
			if envelope.HasMore {
				nextPageCursor = envelope.NextPage
			}
			var items []json.RawMessage
			for _, bucket := range envelope.Data {
				for _, result := range bucket.Results {
					var obj map[string]json.RawMessage
					if jsonErr := json.Unmarshal(result, &obj); jsonErr != nil {
						continue
					}
					startBytes, _ := json.Marshal(bucket.StartingAt)
					endBytes, _ := json.Marshal(bucket.EndingAt)
					obj["_starting_at"] = startBytes
					obj["_ending_at"] = endBytes
					merged, mergeErr := json.Marshal(obj)
					if mergeErr != nil {
						continue
					}
					items = append(items, merged)
				}
			}
			return items, nil
		},
	})
	if err != nil {
		return err
	}

	return apiCollector.Execute()
}
