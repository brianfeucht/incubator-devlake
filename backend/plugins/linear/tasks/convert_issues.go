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
	"github.com/apache/incubator-devlake/plugins/linear/models"
)

var ConvertIssuesMeta = plugin.SubTaskMeta{
	Name:             "ConvertIssues",
	EntryPoint:       ConvertIssues,
	EnabledByDefault: true,
	Description:      "Convert Linear issues into domain layer issues",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{models.LinearIssue{}.TableName()},
	ProductTables:    []string{ticket.Issue{}.TableName(), ticket.BoardIssue{}.TableName()},
}

func ConvertIssues(taskCtx plugin.SubTaskContext) errors.Error {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*LinearTaskData)

	issueIdGen := didgen.NewDomainIdGenerator(&models.LinearIssue{})
	teamIdGen := didgen.NewDomainIdGenerator(&models.LinearTeam{})

	converter, err := helper.NewStatefulDataConverter(&helper.StatefulDataConverterArgs[models.LinearIssue]{
		SubtaskCommonArgs: &helper.SubtaskCommonArgs{
			SubTaskContext: taskCtx,
			Table:          RAW_ISSUES_TABLE,
			Params: models.LinearApiParams{
				ConnectionId: data.Options.ConnectionId,
				TeamId:       data.Options.TeamId,
			},
		},
		Input: func(stateManager *helper.SubtaskStateManager) (dal.Rows, errors.Error) {
			clauses := []dal.Clause{
				dal.From(&models.LinearIssue{}),
				dal.Where("connection_id = ? AND team_id = ?", data.Options.ConnectionId, data.Options.TeamId),
			}
			if stateManager.IsIncremental() {
				if since := stateManager.GetSince(); since != nil {
					clauses = append(clauses, dal.Where("updated_at >= ?", since))
				}
			}
			return db.Cursor(clauses...)
		},
		Convert: func(issue *models.LinearIssue) ([]interface{}, errors.Error) {
			domainIssue := &ticket.Issue{
				DomainEntity:    domainlayer.DomainEntity{Id: issueIdGen.Generate(issue.ConnectionId, issue.TeamId, issue.Id)},
				IssueKey:        issue.Id,
				Title:           issue.Title,
				Description:     issue.Description,
				Url:             issue.Url,
				OriginalStatus:  issue.StateName,
				AssigneeId:      issue.AssigneeId,
				AssigneeName:    issue.AssigneeName,
				OriginalProject: issue.TeamId,
				CreatedDate:     &issue.CreatedAt,
				UpdatedDate:     &issue.UpdatedAt,
				ResolutionDate:  issue.CompletedAt,
				DueDate:         issue.DueDate,
				Priority:        mapLinearPriority(issue.Priority),
			}

			// Map Linear state type to DevLake status
			switch issue.StateType {
			case "completed", "cancelled":
				domainIssue.Status = ticket.DONE
			case "started":
				domainIssue.Status = ticket.IN_PROGRESS
			default:
				domainIssue.Status = ticket.TODO
			}

			domainIssue.Type = ticket.TASK

			boardId := teamIdGen.Generate(issue.ConnectionId, issue.TeamId)
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

func mapLinearPriority(priority int) string {
	switch priority {
	case 1:
		return "HIGHEST" // Urgent
	case 2:
		return "HIGH"
	case 3:
		return "MEDIUM"
	case 4:
		return "LOW"
	default:
		return ""
	}
}
