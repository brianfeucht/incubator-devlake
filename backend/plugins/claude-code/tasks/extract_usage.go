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
	"time"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/claude-code/models"
)

var ExtractUsageReportMeta = plugin.SubTaskMeta{
	Name:             "ExtractUsageReport",
	EntryPoint:       ExtractUsageReport,
	EnabledByDefault: true,
	Description:      "Extract raw Claude usage report data into tool-layer table",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{RAW_USAGE_TABLE},
	ProductTables:    []string{models.ClaudeUsageEntry{}.TableName()},
}

// claudeUsageResult mirrors one result item in the usage report.
type claudeUsageResult struct {
	WorkspaceId                         string `json:"workspace_id"`
	Model                               string `json:"model"`
	AccountId                           string `json:"account_id"`
	ApiKeyId                            string `json:"api_key_id"`
	ServiceAccountId                    string `json:"service_account_id"`
	ServiceTier                         string `json:"service_tier"`
	UncachedInputTokens                 int64  `json:"uncached_input_tokens"`
	OutputTokens                        int64  `json:"output_tokens"`
	CacheReadInputTokens                int64  `json:"cache_read_input_tokens"`
	CacheCreationEphemeral1hInputTokens int64  `json:"cache_creation_ephemeral_1h_input_tokens"`
	CacheCreationEphemeral5mInputTokens int64  `json:"cache_creation_ephemeral_5m_input_tokens"`
	// injected by collector
	StartingAt string `json:"_starting_at"`
	EndingAt   string `json:"_ending_at"`
}

func ExtractUsageReport(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*ClaudeTaskData)

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: claudeApiParams{
				ConnectionId: data.Options.ConnectionId,
				ScopeId:      data.Options.ScopeId,
			},
			Table: RAW_USAGE_TABLE,
		},
		Extract: func(row *helper.RawData) ([]interface{}, errors.Error) {
			var result claudeUsageResult
			if err := json.Unmarshal(row.Data, &result); err != nil {
				return nil, errors.Default.WrapRaw(err)
			}

			startingAt, _ := time.Parse(time.RFC3339, result.StartingAt)
			endingAt, _ := time.Parse(time.RFC3339, result.EndingAt)

			entry := &models.ClaudeUsageEntry{
				ConnectionId:                        data.Options.ConnectionId,
				ScopeId:                             data.Options.ScopeId,
				StartingAt:                          startingAt,
				EndingAt:                            endingAt,
				WorkspaceId:                         result.WorkspaceId,
				Model:                               result.Model,
				AccountId:                           result.AccountId,
				ApiKeyId:                            result.ApiKeyId,
				ServiceAccountId:                    result.ServiceAccountId,
				ServiceTier:                         result.ServiceTier,
				UncachedInputTokens:                 result.UncachedInputTokens,
				OutputTokens:                        result.OutputTokens,
				CacheReadInputTokens:                result.CacheReadInputTokens,
				CacheCreationEphemeral1hInputTokens: result.CacheCreationEphemeral1hInputTokens,
				CacheCreationEphemeral5mInputTokens: result.CacheCreationEphemeral5mInputTokens,
			}
			return []interface{}{entry}, nil
		},
	})
	if err != nil {
		return err
	}
	return extractor.Execute()
}
