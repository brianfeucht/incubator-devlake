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
	"fmt"
	"time"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/domainlayer"
	"github.com/apache/incubator-devlake/core/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/core/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/linear/models"
)

var ConvertSprintsMeta = plugin.SubTaskMeta{
	Name:             "ConvertSprints",
	EntryPoint:       ConvertSprints,
	EnabledByDefault: true,
	Description:      "Convert Linear cycles into domain layer sprints",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{models.LinearCycle{}.TableName()},
	ProductTables:    []string{ticket.Sprint{}.TableName(), ticket.BoardSprint{}.TableName(), ticket.SprintIssue{}.TableName()},
}

func ConvertSprints(taskCtx plugin.SubTaskContext) errors.Error {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*LinearTaskData)

	cycleIdGen := didgen.NewDomainIdGenerator(&models.LinearCycle{})
	teamIdGen := didgen.NewDomainIdGenerator(&models.LinearTeam{})
	issueIdGen := didgen.NewDomainIdGenerator(&models.LinearIssue{})

	boardId := teamIdGen.Generate(data.Options.ConnectionId, data.Options.TeamId)

	converter, err := helper.NewStatefulDataConverter(&helper.StatefulDataConverterArgs[models.LinearCycle]{
		SubtaskCommonArgs: &helper.SubtaskCommonArgs{
			SubTaskContext: taskCtx,
			Table:          RAW_CYCLES_TABLE,
			Params: models.LinearApiParams{
				ConnectionId: data.Options.ConnectionId,
				TeamId:       data.Options.TeamId,
			},
		},
		Input: func(_ *helper.SubtaskStateManager) (dal.Rows, errors.Error) {
			return db.Cursor(
				dal.From(&models.LinearCycle{}),
				dal.Where("connection_id = ? AND team_id = ?", data.Options.ConnectionId, data.Options.TeamId),
			)
		},
		Convert: func(cycle *models.LinearCycle) ([]interface{}, errors.Error) {
			now := time.Now()
			var status string
			if cycle.CompletedAt != nil {
				status = "CLOSED"
			} else if cycle.StartsAt.After(now) {
				status = "FUTURE"
			} else {
				status = "ACTIVE"
			}

			name := fmt.Sprintf("Cycle #%d", cycle.Number)
			if cycle.Name != "" {
				name = fmt.Sprintf("Cycle #%d: %s", cycle.Number, cycle.Name)
			}

			sprintId := cycleIdGen.Generate(cycle.ConnectionId, cycle.TeamId, cycle.Id)
			sprint := &ticket.Sprint{
				DomainEntity:    domainlayer.DomainEntity{Id: sprintId},
				Name:            name,
				Status:          status,
				StartedDate:     &cycle.StartsAt,
				EndedDate:       &cycle.EndsAt,
				CompletedDate:   cycle.CompletedAt,
				OriginalBoardID: boardId,
			}

			boardSprint := &ticket.BoardSprint{
				BoardId:  boardId,
				SprintId: sprintId,
			}

			return []interface{}{sprint, boardSprint}, nil
		},
	})
	if err != nil {
		return err
	}
	if err := converter.Execute(); err != nil {
		return err
	}

	// Link issues that belong to a cycle → SprintIssue
	var issues []models.LinearIssue
	if err := db.All(&issues,
		dal.Where("connection_id = ? AND team_id = ? AND cycle_id != ''", data.Options.ConnectionId, data.Options.TeamId),
	); err != nil {
		return errors.Convert(err)
	}
	for _, issue := range issues {
		sprintId := cycleIdGen.Generate(issue.ConnectionId, issue.TeamId, issue.CycleId)
		issueId := issueIdGen.Generate(issue.ConnectionId, issue.TeamId, issue.Id)
		sprintIssue := &ticket.SprintIssue{
			SprintId: sprintId,
			IssueId:  issueId,
		}
		if err := db.CreateOrUpdate(sprintIssue); err != nil {
			return errors.Convert(err)
		}
	}

	return nil
}
