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
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/plugins/claude-code/models"
)

// ClaudeOptions is (de)serialized from the pipeline task JSON options.
type ClaudeOptions struct {
	ConnectionId uint64 `json:"connectionId"`
	ScopeId      string `json:"scopeId"` // Anthropic workspace ID
	// DaysToSync controls how many days back to fetch on the first (full) sync.
	// Defaults to 30.
	DaysToSync int `json:"daysToSync"`
}

// ClaudeTaskData is made available to all sub-tasks.
type ClaudeTaskData struct {
	Options    *ClaudeOptions
	ApiClient  *helper.ApiAsyncClient
	Connection *models.ClaudeConnection
	Workspace  *models.ClaudeWorkspace
}
