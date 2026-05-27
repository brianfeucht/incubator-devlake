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

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/domainlayer"
	"github.com/apache/incubator-devlake/core/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/core/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/linear/models"
)

var ConvertBoardsMeta = plugin.SubTaskMeta{
	Name:             "ConvertBoards",
	EntryPoint:       ConvertBoards,
	EnabledByDefault: true,
	Description:      "Convert Linear teams into domain layer boards",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{models.LinearTeam{}.TableName()},
	ProductTables:    []string{ticket.Board{}.TableName()},
}

func ConvertBoards(taskCtx plugin.SubTaskContext) errors.Error {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*LinearTaskData)

	teamIdGen := didgen.NewDomainIdGenerator(&models.LinearTeam{})

	converter, err := helper.NewStatefulDataConverter(&helper.StatefulDataConverterArgs[models.LinearTeam]{
		SubtaskCommonArgs: &helper.SubtaskCommonArgs{
			SubTaskContext: taskCtx,
			Table:          RAW_ISSUES_TABLE,
			Params: models.LinearApiParams{
				ConnectionId: data.Options.ConnectionId,
				TeamId:       data.Options.TeamId,
			},
		},
		Input: func(_ *helper.SubtaskStateManager) (dal.Rows, errors.Error) {
			return db.Cursor(
				dal.From(&models.LinearTeam{}),
				dal.Where("connection_id = ? AND id = ?", data.Options.ConnectionId, data.Options.TeamId),
			)
		},
		Convert: func(team *models.LinearTeam) ([]interface{}, errors.Error) {
			board := &ticket.Board{
				DomainEntity: domainlayer.DomainEntity{Id: teamIdGen.Generate(team.ConnectionId, team.Id)},
				Name:         team.Name,
				Url:          fmt.Sprintf("https://linear.app/team/%s/issues", team.Key),
				Type:         "linear",
			}
			return []interface{}{board}, nil
		},
	})
	if err != nil {
		return err
	}
	return converter.Execute()
}
