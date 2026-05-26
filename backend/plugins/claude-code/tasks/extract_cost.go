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
	"strconv"
	"time"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/claude-code/models"
)

var ExtractCostReportMeta = plugin.SubTaskMeta{
	Name:             "ExtractCostReport",
	EntryPoint:       ExtractCostReport,
	EnabledByDefault: true,
	Description:      "Extract raw Claude cost report data into tool-layer table",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{RAW_COST_TABLE},
	ProductTables:    []string{models.ClaudeCostEntry{}.TableName()},
}

// claudeCostResult mirrors one result item in the cost report.
type claudeCostResult struct {
	WorkspaceId   string `json:"workspace_id"`
	Model         string `json:"model"`
	TokenType     string `json:"token_type"`
	CostType      string `json:"cost_type"`
	ContextWindow string `json:"context_window"`
	ServiceTier   string `json:"service_tier"`
	InferenceGeo  string `json:"inference_geo"`
	Amount        string `json:"amount"`
	Currency      string `json:"currency"`
	Description   string `json:"description"`
	// injected by collector
	StartingAt string `json:"_starting_at"`
	EndingAt   string `json:"_ending_at"`
}

func ExtractCostReport(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*ClaudeTaskData)

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: claudeApiParams{
				ConnectionId: data.Options.ConnectionId,
				ScopeId:      data.Options.ScopeId,
			},
			Table: RAW_COST_TABLE,
		},
		Extract: func(row *helper.RawData) ([]interface{}, errors.Error) {
			var result claudeCostResult
			if err := json.Unmarshal(row.Data, &result); err != nil {
				return nil, errors.Default.WrapRaw(err)
			}

			startingAt, _ := time.Parse(time.RFC3339, result.StartingAt)
			endingAt, _ := time.Parse(time.RFC3339, result.EndingAt)
			amount, _ := strconv.ParseFloat(result.Amount, 64)

			entry := &models.ClaudeCostEntry{
				ConnectionId:  data.Options.ConnectionId,
				ScopeId:       data.Options.ScopeId,
				StartingAt:    startingAt,
				EndingAt:      endingAt,
				WorkspaceId:   result.WorkspaceId,
				Model:         result.Model,
				TokenType:     result.TokenType,
				CostType:      result.CostType,
				ContextWindow: result.ContextWindow,
				ServiceTier:   result.ServiceTier,
				InferenceGeo:  result.InferenceGeo,
				Amount:        amount,
				Currency:      result.Currency,
				Description:   result.Description,
			}
			return []interface{}{entry}, nil
		},
	})
	if err != nil {
		return err
	}
	return extractor.Execute()
}
