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
	"github.com/go-playground/validator/v10"

	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/sentry/models"
)

var (
	basicRes         context.BasicRes
	vld              *validator.Validate
	connectionHelper *helper.ConnectionApiHelper
	dsHelper         *helper.DsHelper[models.SentryConnection, models.SentryProject, models.SentryScopeConfig]
	raProxy          *helper.DsRemoteApiProxyHelper[models.SentryConnection]
	raScopeList      *helper.DsRemoteApiScopeListHelper[models.SentryConnection, models.SentryProject, SentryRemotePagination]
	raScopeSearch    *helper.DsRemoteApiScopeSearchHelper[models.SentryConnection, models.SentryProject]
)

// Init stores basic resources and wires shared helpers for all API handlers.
func Init(br context.BasicRes, meta plugin.PluginMeta) {
	basicRes = br
	vld = validator.New()
	connectionHelper = helper.NewConnectionHelper(basicRes, vld, meta.Name())
	dsHelper = helper.NewDataSourceHelper[
		models.SentryConnection, models.SentryProject, models.SentryScopeConfig,
	](
		basicRes,
		meta.Name(),
		[]string{"id", "org_slug", "slug"},
		func(c models.SentryConnection) models.SentryConnection {
			c.Normalize()
			return c.Sanitize()
		},
		func(s models.SentryProject) models.SentryProject { return s },
		nil,
	)
	raProxy = helper.NewDsRemoteApiProxyHelper[models.SentryConnection](dsHelper.ConnApi.ModelApiHelper)
	raScopeList = helper.NewDsRemoteApiScopeListHelper[models.SentryConnection, models.SentryProject, SentryRemotePagination](raProxy, listSentryRemoteScopes)
	raScopeSearch = helper.NewDsRemoteApiScopeSearchHelper[models.SentryConnection, models.SentryProject](raProxy, searchSentryRemoteScopes)
}
