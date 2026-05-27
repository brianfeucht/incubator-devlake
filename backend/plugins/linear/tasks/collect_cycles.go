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

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/linear/models"
)

const RAW_CYCLES_TABLE = "linear_cycles"

var CollectCyclesMeta = plugin.SubTaskMeta{
	Name:             "CollectCycles",
	EntryPoint:       CollectCycles,
	EnabledByDefault: true,
	Description:      "Collect Linear cycles (sprints) for a team",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{},
	ProductTables:    []string{RAW_CYCLES_TABLE},
}

func CollectCycles(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*LinearTaskData)

	apiCollector, err := helper.NewStatefulApiCollector(helper.RawDataSubTaskArgs{
		Ctx: taskCtx,
		Params: models.LinearApiParams{
			ConnectionId: data.Options.ConnectionId,
			TeamId:       data.Options.TeamId,
		},
		Table: RAW_CYCLES_TABLE,
	})
	if err != nil {
		return err
	}

	var nextCursor string

	err = apiCollector.InitCollector(helper.ApiCollectorArgs{
		ApiClient:   data.ApiClient,
		PageSize:    50,
		Method:      http.MethodPost,
		UrlTemplate: "graphql",
		RequestBody: func(reqData *helper.RequestData) map[string]interface{} {
			cursor := ""
			if reqData.CustomData != nil {
				cursor = reqData.CustomData.(string)
			}
			afterClause := ""
			if cursor != "" {
				afterClause = fmt.Sprintf(`, after: "%s"`, cursor)
			}
			query := fmt.Sprintf(`{ team(id: "%s") { cycles(first: 50%s) { nodes { id number name startsAt endsAt completedAt createdAt updatedAt } pageInfo { hasNextPage endCursor } } } }`,
				data.Options.TeamId, afterClause)
			return map[string]interface{}{"query": query}
		},
		GetNextPageCustomData: func(reqData *helper.RequestData, res *http.Response) (interface{}, errors.Error) {
			if nextCursor == "" {
				return nil, nil
			}
			c := nextCursor
			nextCursor = ""
			return c, nil
		},
		ResponseParser: func(res *http.Response) ([]json.RawMessage, errors.Error) {
			type pageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			}
			type cyclesConn struct {
				Nodes    []json.RawMessage `json:"nodes"`
				PageInfo pageInfo          `json:"pageInfo"`
			}
			type teamData struct {
				Cycles cyclesConn `json:"cycles"`
			}
			type dataWrapper struct {
				Team teamData `json:"team"`
			}
			type gqlResp struct {
				Data dataWrapper `json:"data"`
			}
			var resp gqlResp
			if err := helper.UnmarshalResponse(res, &resp); err != nil {
				return nil, err
			}
			if resp.Data.Team.Cycles.PageInfo.HasNextPage {
				nextCursor = resp.Data.Team.Cycles.PageInfo.EndCursor
			}
			return resp.Data.Team.Cycles.Nodes, nil
		},
	})
	if err != nil {
		return err
	}
	return apiCollector.Execute()
}
