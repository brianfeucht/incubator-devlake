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
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/domainlayer"
	"github.com/apache/incubator-devlake/core/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/core/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/sentry/models"
)

var ConvertIssuesMeta = plugin.SubTaskMeta{
	Name:             "ConvertIssues",
	EntryPoint:       ConvertIssues,
	EnabledByDefault: true,
	Description:      "Convert tool layer Sentry issues into domain layer issues",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{models.SentryIssue{}.TableName()},
	ProductTables:    []string{ticket.Issue{}.TableName(), ticket.BoardIssue{}.TableName()},
}

func ConvertIssues(taskCtx plugin.SubTaskContext) errors.Error {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*SentryTaskData)

	issueIdGen := didgen.NewDomainIdGenerator(&models.SentryIssue{})
	boardIdGen := didgen.NewDomainIdGenerator(&models.SentryProject{})

	converter, err := helper.NewStatefulDataConverter(&helper.StatefulDataConverterArgs[models.SentryIssue]{
		SubtaskCommonArgs: &helper.SubtaskCommonArgs{
			SubTaskContext: taskCtx,
			Table:          RAW_ISSUE_TABLE,
			Params: SentryApiParams{
				ConnectionId: data.Options.ConnectionId,
				ScopeId:      data.Options.ScopeId,
			},
		},
		Input: func(stateManager *helper.SubtaskStateManager) (dal.Rows, errors.Error) {
			clauses := []dal.Clause{
				dal.From(&models.SentryIssue{}),
				dal.Where("connection_id = ? AND scope_id = ?",
					data.Options.ConnectionId, data.Options.ScopeId),
			}
			if stateManager.IsIncremental() {
				since := stateManager.GetSince()
				if since != nil {
					clauses = append(clauses, dal.Where("updated_at >= ?", since))
				}
			}
			return db.Cursor(clauses...)
		},
		Convert: func(issue *models.SentryIssue) ([]interface{}, errors.Error) {
			domainIssue := &ticket.Issue{
				DomainEntity:   domainlayer.DomainEntity{Id: issueIdGen.Generate(issue.ConnectionId, issue.ScopeId, issue.IssueId)},
				IssueKey:       issue.ShortId,
				Title:          issue.Title,
				Description:    issue.Culprit,
				Url:            issue.Permalink,
				OriginalStatus: issue.Status,
				AssigneeName:   issue.AssigneeName,
				AssigneeId:     issue.AssigneeEmail,
				OriginalProject: data.Options.ScopeId,
				CreatedDate:    issue.FirstSeen,
				UpdatedDate:    issue.LastSeen,
			}

			// Map Sentry status to DevLake status constants
			switch issue.Status {
			case "resolved", "ignored":
				domainIssue.Status = ticket.DONE
			case "unresolved":
				domainIssue.Status = ticket.TODO
			default:
				domainIssue.Status = ticket.OTHER
			}

			// Map Sentry issue category to DevLake type constants
			switch issue.IssueCategory {
			case "error":
				domainIssue.Type = ticket.BUG
			case "performance":
				domainIssue.Type = ticket.TASK
			default:
				domainIssue.Type = ticket.TASK
			}

			// Map Sentry level to DevLake severity
			switch issue.Level {
			case "fatal", "error":
				domainIssue.Severity = "HIGH"
			case "warning":
				domainIssue.Severity = "MEDIUM"
			default:
				domainIssue.Severity = "LOW"
			}

			boardId := boardIdGen.Generate(issue.ConnectionId, issue.ScopeId)
			boardIssue := &ticket.BoardIssue{
				BoardId: boardId,
				IssueId: domainIssue.Id,
			}

			return []interface{}{domainIssue, boardIssue}, nil
		},
	})
	if err != nil {
		return err
	}
	return converter.Execute()
}
