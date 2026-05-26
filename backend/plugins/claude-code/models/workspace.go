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

package models

import (
	"github.com/apache/incubator-devlake/core/models/common"
	"github.com/apache/incubator-devlake/core/plugin"
)

// ClaudeWorkspace represents an Anthropic workspace used as a DevLake data scope.
// The scope ID is the Anthropic workspace ID (e.g. "wrkspc_01JwQvzr7rXLA5AGx3HKfFUJ").
type ClaudeWorkspace struct {
	common.Scope  `mapstructure:",squash"`
	Id            string `json:"id" mapstructure:"id" gorm:"primaryKey;type:varchar(255)"`
	Name          string `json:"name" mapstructure:"name" gorm:"type:varchar(255)"`
	DisplayName   string `json:"displayName" mapstructure:"displayName" gorm:"type:varchar(255)"`
}

func (ClaudeWorkspace) TableName() string {
	return "_tool_claude_workspaces"
}

func (w ClaudeWorkspace) ScopeId() string {
	return w.Id
}

func (w ClaudeWorkspace) ScopeName() string {
	if w.DisplayName != "" {
		return w.DisplayName
	}
	if w.Name != "" {
		return w.Name
	}
	return w.Id
}

func (w ClaudeWorkspace) ScopeFullName() string {
	return w.Id
}

func (w ClaudeWorkspace) ScopeParams() interface{} {
	return &ClaudeWorkspaceParams{
		ConnectionId: w.ConnectionId,
		ScopeId:      w.Id,
	}
}

type ClaudeWorkspaceParams struct {
	ConnectionId uint64 `json:"connectionId"`
	ScopeId      string `json:"scopeId"`
}

var _ plugin.ToolLayerScope = (*ClaudeWorkspace)(nil)
