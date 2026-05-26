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
	"github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/sentry/models"
)

// SentryOptions is (de)serialized from the pipeline task JSON options.
type SentryOptions struct {
	ConnectionId uint64 `json:"connectionId"`
	ScopeId      string `json:"scopeId"` // "<orgSlug>/<projectSlug>"
	// OrgSlug and ProjectSlug are derived from ScopeId at prep time.
	OrgSlug     string `json:"orgSlug"`
	ProjectSlug string `json:"projectSlug"`
}

// SentryTaskData is made available to all sub-tasks.
type SentryTaskData struct {
	Options    *SentryOptions
	ApiClient  *api.ApiAsyncClient
	Connection *models.SentryConnection
	Project    *models.SentryProject
}
