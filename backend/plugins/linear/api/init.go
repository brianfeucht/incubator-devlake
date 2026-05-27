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
	"github.com/apache/incubator-devlake/plugins/linear/models"
)

var (
	basicRes         context.BasicRes
	vld              *validator.Validate
	connectionHelper *helper.ConnectionApiHelper
	dsHelper         *helper.DsHelper[models.LinearConnection, models.LinearTeam, models.LinearScopeConfig]
	raProxy          *helper.DsRemoteApiProxyHelper[models.LinearConnection]
	raScopeList      *helper.DsRemoteApiScopeListHelper[models.LinearConnection, models.LinearTeam, LinearRemotePagination]
	raScopeSearch    *helper.DsRemoteApiScopeSearchHelper[models.LinearConnection, models.LinearTeam]
)

// Init stores basic resources and wires shared helpers for all API handlers.
func Init(br context.BasicRes, meta plugin.PluginMeta) {
	basicRes = br
	vld = validator.New()
	connectionHelper = helper.NewConnectionHelper(basicRes, vld, meta.Name())
	dsHelper = helper.NewDataSourceHelper[
		models.LinearConnection, models.LinearTeam, models.LinearScopeConfig,
	](
		basicRes,
		meta.Name(),
		[]string{"id", "name"},
		func(c models.LinearConnection) models.LinearConnection {
			c.Normalize()
			return c.Sanitize()
		},
		func(s models.LinearTeam) models.LinearTeam { return s },
		nil,
	)
	raProxy = helper.NewDsRemoteApiProxyHelper[models.LinearConnection](dsHelper.ConnApi.ModelApiHelper)
	raScopeList = helper.NewDsRemoteApiScopeListHelper[models.LinearConnection, models.LinearTeam, LinearRemotePagination](raProxy, listLinearRemoteScopes)
	raScopeSearch = helper.NewDsRemoteApiScopeSearchHelper[models.LinearConnection, models.LinearTeam](raProxy, searchLinearRemoteScopes)
}
