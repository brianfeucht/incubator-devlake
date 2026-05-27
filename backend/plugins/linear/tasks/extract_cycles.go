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

var ExtractCyclesMeta = plugin.SubTaskMeta{
	Name:             "ExtractCycles",
	EntryPoint:       ExtractCycles,
	EnabledByDefault: true,
	Description:      "Extract raw Linear cycle data into the tool layer",
	DomainTypes:      []string{plugin.DOMAIN_TYPE_TICKET},
	DependencyTables: []string{RAW_CYCLES_TABLE},
	ProductTables:    []string{models.LinearCycle{}.TableName()},
}

type linearCycleResponse struct {
	Id          string     `json:"id"`
	Number      int        `json:"number"`
	Name        string     `json:"name"`
	StartsAt    time.Time  `json:"startsAt"`
	EndsAt      time.Time  `json:"endsAt"`
	CompletedAt *time.Time `json:"completedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func ExtractCycles(taskCtx plugin.SubTaskContext) errors.Error {
	data := taskCtx.GetData().(*LinearTaskData)

	extractor, err := helper.NewStatefulApiExtractor(&helper.StatefulApiExtractorArgs[linearCycleResponse]{
		SubtaskCommonArgs: &helper.SubtaskCommonArgs{
			SubTaskContext: taskCtx,
			Params: models.LinearApiParams{
				ConnectionId: data.Options.ConnectionId,
				TeamId:       data.Options.TeamId,
			},
			Table: RAW_CYCLES_TABLE,
		},
		Extract: func(row *linearCycleResponse, _ *helper.RawData) ([]interface{}, errors.Error) {
			cycle := &models.LinearCycle{
				ConnectionId: data.Options.ConnectionId,
				TeamId:       data.Options.TeamId,
				Id:           row.Id,
				Number:       row.Number,
				Name:         row.Name,
				StartsAt:     row.StartsAt,
				EndsAt:       row.EndsAt,
				CompletedAt:  row.CompletedAt,
				CreatedAt:    row.CreatedAt,
				UpdatedAt:    row.UpdatedAt,
			}
			return []interface{}{cycle}, nil
		},
	})
	if err != nil {
		return err
	}
	return extractor.Execute()
}
