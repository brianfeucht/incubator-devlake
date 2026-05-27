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

package impl

import (
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	coreModels "github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/linear/api"
	"github.com/apache/incubator-devlake/plugins/linear/models"
	"github.com/apache/incubator-devlake/plugins/linear/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/linear/tasks"
)

var _ interface {
	plugin.PluginMeta
	plugin.PluginInit
	plugin.PluginTask
	plugin.PluginApi
	plugin.PluginModel
	plugin.PluginSource
	plugin.DataSourcePluginBlueprintV200
	plugin.PluginMigration
	plugin.CloseablePluginTask
} = (*Linear)(nil)

// Linear is the DevLake plugin entrypoint.
type Linear struct{}

func (p Linear) Init(basicRes context.BasicRes) errors.Error {
	api.Init(basicRes, p)
	return nil
}

func (p Linear) Name() string { return "linear" }

func (p Linear) Description() string {
	return "Collect Linear issues, teams, and cycles"
}

func (p Linear) Connection() dal.Tabler       { return &models.LinearConnection{} }
func (p Linear) Scope() plugin.ToolLayerScope { return &models.LinearTeam{} }
func (p Linear) ScopeConfig() dal.Tabler      { return &models.LinearScopeConfig{} }

func (p Linear) GetTablesInfo() []dal.Tabler { return models.GetTablesInfo() }

func (p Linear) SubTaskMetas() []plugin.SubTaskMeta { return tasks.GetSubTaskMetas() }

func (p Linear) PrepareTaskData(taskCtx plugin.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	var op tasks.LinearOptions
	if err := helper.Decode(options, &op, nil); err != nil {
		return nil, err
	}

	connectionHelper := helper.NewConnectionHelper(taskCtx, nil, p.Name())
	connection := &models.LinearConnection{}
	if err := connectionHelper.FirstById(connection, op.ConnectionId); err != nil {
		return nil, err
	}
	connection.Normalize()

	db := taskCtx.GetDal()
	team := &models.LinearTeam{}
	if err := db.First(team, dal.Where("connection_id = ? AND id = ?", op.ConnectionId, op.TeamId)); err != nil {
		taskCtx.GetLogger().Warn(errors.Convert(err), "team scope not found in db; continuing without it")
		team = nil
	}

	asyncClient, err := tasks.CreateAsyncApiClient(taskCtx, connection)
	if err != nil {
		return nil, err
	}

	return &tasks.LinearTaskData{
		Options:   &op,
		ApiClient: asyncClient,
		Team:      team,
	}, nil
}

func (p Linear) ApiResources() map[string]map[string]plugin.ApiResourceHandler {
	return map[string]map[string]plugin.ApiResourceHandler{
		"test": {
			"POST": api.TestConnection,
		},
		"connections": {
			"POST": api.PostConnections,
			"GET":  api.ListConnections,
		},
		"connections/:connectionId": {
			"GET":    api.GetConnection,
			"PATCH":  api.PatchConnection,
			"DELETE": api.DeleteConnection,
		},
		"connections/:connectionId/test": {
			"POST": api.TestExistingConnection,
		},
		"connections/:connectionId/scopes": {
			"GET": api.GetScopeList,
			"PUT": api.PutScopes,
		},
		"connections/:connectionId/scopes/:scopeId": {
			"GET":    api.GetScope,
			"PATCH":  api.PatchScope,
			"DELETE": api.DeleteScope,
		},
		"connections/:connectionId/scopes/:scopeId/latest-sync-state": {
			"GET": api.GetScopeLatestSyncState,
		},
		"connections/:connectionId/remote-scopes": {
			"GET": api.RemoteScopes,
		},
		"connections/:connectionId/search-remote-scopes": {
			"GET": api.SearchRemoteScopes,
		},
		"connections/:connectionId/proxy/rest/*path": {
			"GET": api.Proxy,
		},
		"connections/:connectionId/scope-configs": {
			"POST": api.PostScopeConfig,
			"GET":  api.GetScopeConfigList,
		},
		"connections/:connectionId/scope-configs/:scopeConfigId": {
			"GET":    api.GetScopeConfig,
			"PATCH":  api.PatchScopeConfig,
			"DELETE": api.DeleteScopeConfig,
		},
	}
}

func (p Linear) MakeDataSourcePipelinePlanV200(
	connectionId uint64,
	scopes []*coreModels.BlueprintScope,
) (coreModels.PipelinePlan, []plugin.Scope, errors.Error) {
	return api.MakeDataSourcePipelinePlanV200(p.SubTaskMetas(), connectionId, scopes)
}

func (p Linear) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/linear"
}

func (p Linear) MigrationScripts() []plugin.MigrationScript {
	return migrationscripts.All()
}

func (p Linear) Close(taskCtx plugin.TaskContext) errors.Error {
	if data, ok := taskCtx.GetData().(*tasks.LinearTaskData); ok && data != nil {
		if data.ApiClient != nil {
			data.ApiClient.WaitAsync()
		}
	}
	return nil
}
