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
	"github.com/apache/incubator-devlake/plugins/claude-code/api"
	"github.com/apache/incubator-devlake/plugins/claude-code/models"
	"github.com/apache/incubator-devlake/plugins/claude-code/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/claude-code/tasks"
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
} = (*ClaudeCode)(nil)

// ClaudeCode is the DevLake plugin entrypoint.
type ClaudeCode struct{}

func (p ClaudeCode) Init(basicRes context.BasicRes) errors.Error {
	api.Init(basicRes, p)
	return nil
}

func (p ClaudeCode) Name() string { return "claude-code" }

func (p ClaudeCode) Description() string {
	return "Collect Anthropic Claude Code cost and token usage data via the Admin API"
}

func (p ClaudeCode) Connection() dal.Tabler { return &models.ClaudeConnection{} }
func (p ClaudeCode) Scope() plugin.ToolLayerScope { return &models.ClaudeWorkspace{} }
func (p ClaudeCode) ScopeConfig() dal.Tabler { return &models.ClaudeScopeConfig{} }

func (p ClaudeCode) GetTablesInfo() []dal.Tabler { return models.GetTablesInfo() }

func (p ClaudeCode) SubTaskMetas() []plugin.SubTaskMeta { return tasks.GetSubTaskMetas() }

func (p ClaudeCode) PrepareTaskData(taskCtx plugin.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	var op tasks.ClaudeOptions
	if err := helper.Decode(options, &op, nil); err != nil {
		return nil, err
	}

	connectionHelper := helper.NewConnectionHelper(taskCtx, nil, p.Name())
	connection := &models.ClaudeConnection{}
	if err := connectionHelper.FirstById(connection, op.ConnectionId); err != nil {
		return nil, err
	}
	connection.Normalize()

	// Load the saved workspace scope for reference.
	db := taskCtx.GetDal()
	workspace := &models.ClaudeWorkspace{}
	if err := db.First(workspace, dal.Where("connection_id = ? AND id = ?", op.ConnectionId, op.ScopeId)); err != nil {
		taskCtx.GetLogger().Warn(errors.Convert(err), "workspace scope not found in db; continuing without it")
		workspace = nil
	}

	asyncClient, err := tasks.CreateAsyncApiClient(taskCtx, connection)
	if err != nil {
		return nil, err
	}

	return &tasks.ClaudeTaskData{
		Options:    &op,
		ApiClient:  asyncClient,
		Connection: connection,
		Workspace:  workspace,
	}, nil
}

func (p ClaudeCode) ApiResources() map[string]map[string]plugin.ApiResourceHandler {
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

func (p ClaudeCode) MakeDataSourcePipelinePlanV200(
	connectionId uint64,
	scopes []*coreModels.BlueprintScope,
) (coreModels.PipelinePlan, []plugin.Scope, errors.Error) {
	return api.MakeDataSourcePipelinePlanV200(p.SubTaskMetas(), connectionId, scopes)
}

func (p ClaudeCode) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/claude-code"
}

func (p ClaudeCode) MigrationScripts() []plugin.MigrationScript {
	return migrationscripts.All()
}

func (p ClaudeCode) Close(taskCtx plugin.TaskContext) errors.Error {
	if taskData, ok := taskCtx.GetData().(*tasks.ClaudeTaskData); ok && taskData != nil {
		if taskData.ApiClient != nil {
			taskData.ApiClient.WaitAsync()
		}
	}
	return nil
}
