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

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/common"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/sentry/models"
)

var ExtractIssuesMeta = plugin.SubTaskMeta{
	Name:             "ExtractIssues",
	EntryPoint:       ExtractIssues,
	EnabledByDefault: true,
	Description:      "Extract raw Sentry issue data into the tool layer table",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{RAW_ISSUE_TABLE},
	ProductTables:    []string{models.SentryIssue{}.TableName()},
}

// SentryIssueResponse mirrors the Sentry API issue object.
type SentryIssueResponse struct {
	ID            string          `json:"id"`
	ShortID       string          `json:"shortId"`
	Title         string          `json:"title"`
	Culprit       string          `json:"culprit"`
	Permalink     string          `json:"permalink"`
	Level         string          `json:"level"`
	Status        string          `json:"status"`
	Platform      string          `json:"platform"`
	IssueType     string          `json:"issueType"`
	IssueCategory string          `json:"issueCategory"`
	Priority      string          `json:"priority"`
	Count         string          `json:"count"`
	UserCount     int             `json:"userCount"`
	FirstSeen     *common.Iso8601Time `json:"firstSeen"`
	LastSeen      *common.Iso8601Time `json:"lastSeen"`
	Assignee      json.RawMessage `json:"assignee"`
	Project       struct {
		Slug string `json:"slug"`
	} `json:"project"`
}

// sentryAssignee is a partial parse of the assignee field.
type sentryAssignee struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func ExtractIssues(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*SentryTaskData)

	extractor, err := helper.NewStatefulApiExtractor(&helper.StatefulApiExtractorArgs[SentryIssueResponse]{
		SubtaskCommonArgs: &helper.SubtaskCommonArgs{
			SubTaskContext: taskCtx,
			Params: SentryApiParams{
				ConnectionId: data.Options.ConnectionId,
				ScopeId:      data.Options.ScopeId,
			},
			Table: RAW_ISSUE_TABLE,
		},
		Extract: func(row *SentryIssueResponse, rawData *helper.RawData) ([]interface{}, errors.Error) {
			issue := &models.SentryIssue{
				ConnectionId:  data.Options.ConnectionId,
				ScopeId:       data.Options.ScopeId,
				IssueId:       row.ID,
				ShortId:       row.ShortID,
				Title:         row.Title,
				Culprit:       row.Culprit,
				Permalink:     row.Permalink,
				Level:         row.Level,
				Status:        row.Status,
				Platform:      row.Platform,
				IssueType:     row.IssueType,
				IssueCategory: row.IssueCategory,
				Priority:      row.Priority,
				Count:         row.Count,
				UserCount:     row.UserCount,
				OrgSlug:       data.Options.OrgSlug,
				ProjectSlug:   data.Options.ProjectSlug,
			}

			if row.FirstSeen != nil {
				t := row.FirstSeen.ToTime()
				issue.FirstSeen = &t
			}
			if row.LastSeen != nil {
				t := row.LastSeen.ToTime()
				issue.LastSeen = &t
			}

			// Assignee is a union type in the Sentry API — try to parse it.
			if len(row.Assignee) > 0 && string(row.Assignee) != "null" {
				var a sentryAssignee
				if jsonErr := json.Unmarshal(row.Assignee, &a); jsonErr == nil {
					issue.AssigneeName = a.Name
					issue.AssigneeEmail = a.Email
				}
			}

			return []interface{}{issue}, nil
		},
	})
	if err != nil {
		return err
	}
	return extractor.Execute()
}
