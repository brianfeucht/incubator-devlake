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

package api

import (
	"github.com/apache/incubator-devlake/core/errors"
	coreModels "github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/helpers/srvhelper"
	"github.com/apache/incubator-devlake/plugins/sentry/models"
	"github.com/apache/incubator-devlake/plugins/sentry/tasks"
)

// MakeDataSourcePipelinePlanV200 generates the pipeline plan for blueprint v2.0.0.
func MakeDataSourcePipelinePlanV200(
	subtaskMetas []plugin.SubTaskMeta,
	connectionId uint64,
	bpScopes []*coreModels.BlueprintScope,
) (coreModels.PipelinePlan, []plugin.Scope, errors.Error) {
	_, err := dsHelper.ConnSrv.FindByPk(connectionId)
	if err != nil {
		return nil, nil, err
	}
	scopeDetails, err := dsHelper.ScopeSrv.MapScopeDetails(connectionId, bpScopes)
	if err != nil {
		return nil, nil, err
	}

	plan, scopes, err := makeDataSourcePipelinePlanV200(subtaskMetas, scopeDetails)
	if err != nil {
		return nil, nil, err
	}
	return plan, scopes, nil
}

func makeDataSourcePipelinePlanV200(
	subtaskMetas []plugin.SubTaskMeta,
	scopeDetails []*srvhelper.ScopeDetail[models.SentryProject, models.SentryScopeConfig],
) (coreModels.PipelinePlan, []plugin.Scope, errors.Error) {
	plan := make(coreModels.PipelinePlan, len(scopeDetails))
	var domainScopes []plugin.Scope

	for i, scopeDetail := range scopeDetails {
		stage := plan[i]
		if stage == nil {
			stage = coreModels.PipelineStage{}
		}

		scope := scopeDetail.Scope
		task, err := helper.MakePipelinePlanTask(
			"sentry",
			subtaskMetas,
			nil,
			tasks.SentryOptions{
				ConnectionId: scope.ConnectionId,
				ScopeId:      scope.Id,
			},
		)
		if err != nil {
			return nil, nil, err
		}

		stage = append(stage, task)
		plan[i] = stage
	}

	return plan, domainScopes, nil
}
