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
	"strings"

	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	coreModels "github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/sentry/api"
	"github.com/apache/incubator-devlake/plugins/sentry/models"
	"github.com/apache/incubator-devlake/plugins/sentry/models/migrationscripts"
	"github.com/apache/incubator-devlake/plugins/sentry/tasks"
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
} = (*Sentry)(nil)

// Sentry is the DevLake plugin entrypoint.
type Sentry struct{}

func (p Sentry) Init(basicRes context.BasicRes) errors.Error {
	api.Init(basicRes, p)
	return nil
}

func (p Sentry) Name() string { return "sentry" }

func (p Sentry) Description() string {
	return "Collect Sentry issues, projects, and release data"
}

func (p Sentry) Connection() dal.Tabler { return &models.SentryConnection{} }
func (p Sentry) Scope() plugin.ToolLayerScope { return &models.SentryProject{} }
func (p Sentry) ScopeConfig() dal.Tabler { return &models.SentryScopeConfig{} }

func (p Sentry) GetTablesInfo() []dal.Tabler { return models.GetTablesInfo() }

func (p Sentry) SubTaskMetas() []plugin.SubTaskMeta { return tasks.GetSubTaskMetas() }

func (p Sentry) PrepareTaskData(taskCtx plugin.TaskContext, options map[string]interface{}) (interface{}, errors.Error) {
	var op tasks.SentryOptions
	if err := helper.Decode(options, &op, nil); err != nil {
		return nil, err
	}

	connectionHelper := helper.NewConnectionHelper(taskCtx, nil, p.Name())
	connection := &models.SentryConnection{}
	if err := connectionHelper.FirstById(connection, op.ConnectionId); err != nil {
		return nil, err
	}
	connection.Normalize()

	// Derive OrgSlug and ProjectSlug from ScopeId ("<orgSlug>/<projectSlug>")
	if op.OrgSlug == "" || op.ProjectSlug == "" {
		parts := strings.SplitN(op.ScopeId, "/", 2)
		if len(parts) == 2 {
			op.OrgSlug = parts[0]
			op.ProjectSlug = parts[1]
		}
	}

	// Load the saved project scope for reference.
	db := taskCtx.GetDal()
	project := &models.SentryProject{}
	if err := db.First(project, dal.Where("connection_id = ? AND id = ?", op.ConnectionId, op.ScopeId)); err != nil {
		// Non-fatal: project may not exist yet (first run before scope was saved)
		taskCtx.GetLogger().Warn(errors.Convert(err), "project scope not found in db; continuing without it")
		project = nil
	}

	asyncClient, err := tasks.CreateAsyncApiClient(taskCtx, connection)
	if err != nil {
		return nil, err
	}

	return &tasks.SentryTaskData{
		Options:    &op,
		ApiClient:  asyncClient,
		Connection: connection,
		Project:    project,
	}, nil
}

func (p Sentry) ApiResources() map[string]map[string]plugin.ApiResourceHandler {
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

func (p Sentry) MakeDataSourcePipelinePlanV200(
	connectionId uint64,
	scopes []*coreModels.BlueprintScope,
) (coreModels.PipelinePlan, []plugin.Scope, errors.Error) {
	return api.MakeDataSourcePipelinePlanV200(p.SubTaskMetas(), connectionId, scopes)
}

func (p Sentry) RootPkgPath() string {
	return "github.com/apache/incubator-devlake/plugins/sentry"
}

func (p Sentry) MigrationScripts() []plugin.MigrationScript {
	return migrationscripts.All()
}

func (p Sentry) Close(taskCtx plugin.TaskContext) errors.Error {
	if asyncClient, ok := taskCtx.GetData().(*tasks.SentryTaskData); ok && asyncClient != nil {
		if asyncClient.ApiClient != nil {
			asyncClient.ApiClient.WaitAsync()
		}
	}
	return nil
}
