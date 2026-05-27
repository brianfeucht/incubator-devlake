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
	"time"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/linear/models"
)

var ExtractIssuesMeta = plugin.SubTaskMeta{
	Name:             "ExtractIssues",
	EntryPoint:       ExtractIssues,
	EnabledByDefault: true,
	Description:      "Extract raw Linear issue data into the tool layer",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{RAW_ISSUES_TABLE},
	ProductTables:    []string{models.LinearIssue{}.TableName()},
}

// linearIssueResponse mirrors the GraphQL issue node fields.
type linearIssueResponse struct {
	Id            string     `json:"id"`
	Number        int        `json:"number"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Priority      int        `json:"priority"`
	PriorityLabel string     `json:"priorityLabel"`
	Estimate      *float64   `json:"estimate"`
	Url           string     `json:"url"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	CompletedAt   *time.Time `json:"completedAt"`
	CanceledAt    *time.Time `json:"canceledAt"`
	StartedAt     *time.Time `json:"startedAt"`
	DueDate       *time.Time `json:"dueDate"`
	State         struct {
		Id   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"state"`
	Assignee *struct {
		Id    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"assignee"`
	Creator *struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"creator"`
	Cycle *struct {
		Id string `json:"id"`
	} `json:"cycle"`
	Team struct {
		Id string `json:"id"`
	} `json:"team"`
}

func ExtractIssues(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*LinearTaskData)

	extractor, err := helper.NewStatefulApiExtractor(&helper.StatefulApiExtractorArgs[linearIssueResponse]{
		SubtaskCommonArgs: &helper.SubtaskCommonArgs{
			SubTaskContext: taskCtx,
			Params: models.LinearApiParams{
				ConnectionId: data.Options.ConnectionId,
				TeamId:       data.Options.TeamId,
			},
			Table: RAW_ISSUES_TABLE,
		},
		Extract: func(row *linearIssueResponse, _ *helper.RawData) ([]interface{}, errors.Error) {
			issue := &models.LinearIssue{
				ConnectionId:  data.Options.ConnectionId,
				TeamId:        data.Options.TeamId,
				Id:            row.Id,
				Number:        row.Number,
				Title:         row.Title,
				Description:   row.Description,
				Priority:      row.Priority,
				PriorityLabel: row.PriorityLabel,
				Estimate:      row.Estimate,
				Url:           row.Url,
				StateId:       row.State.Id,
				StateName:     row.State.Name,
				StateType:     row.State.Type,
				CompletedAt:   row.CompletedAt,
				CanceledAt:    row.CanceledAt,
				StartedAt:     row.StartedAt,
				DueDate:       row.DueDate,
				CreatedAt:     row.CreatedAt,
				UpdatedAt:     row.UpdatedAt,
			}
			if row.Assignee != nil {
				issue.AssigneeId = row.Assignee.Id
				issue.AssigneeName = row.Assignee.Name
				issue.AssigneeEmail = row.Assignee.Email
			}
			if row.Creator != nil {
				issue.CreatorId = row.Creator.Id
				issue.CreatorName = row.Creator.Name
			}
			if row.Cycle != nil {
				issue.CycleId = row.Cycle.Id
			}
			return []interface{}{issue}, nil
		},
	})
	if err != nil {
		return err
	}
	return extractor.Execute()
}
