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
	"time"

	"github.com/apache/incubator-devlake/core/models/common"
)

// ClaudeUsageEntry holds one usage line-item returned by the Anthropic usage_report/messages API.
type ClaudeUsageEntry struct {
	common.NoPKModel
	ConnectionId                         uint64    `gorm:"primaryKey"`
	ScopeId                              string    `gorm:"primaryKey;type:varchar(255)"` // workspace ID
	StartingAt                           time.Time `gorm:"primaryKey"`
	EndingAt                             time.Time
	WorkspaceId                          string `gorm:"type:varchar(255)"`
	Model                                string `gorm:"type:varchar(255)"`
	AccountId                            string `gorm:"type:varchar(255)"` // user/member ID
	ApiKeyId                             string `gorm:"type:varchar(255)"`
	ServiceAccountId                     string `gorm:"type:varchar(255)"`
	ServiceTier                          string `gorm:"type:varchar(50)"`
	UncachedInputTokens                  int64
	OutputTokens                         int64
	CacheReadInputTokens                 int64
	CacheCreationEphemeral1hInputTokens  int64
	CacheCreationEphemeral5mInputTokens  int64
}

func (ClaudeUsageEntry) TableName() string {
	return "_tool_claude_usage_entries"
}
