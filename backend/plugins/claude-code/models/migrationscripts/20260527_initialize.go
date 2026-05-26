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

package migrationscripts

import (
	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/models/migrationscripts/archived"
	"github.com/apache/incubator-devlake/helpers/migrationhelper"
	"time"
)

type addClaudeInitialTables struct{}

func (s *addClaudeInitialTables) Up(basicRes context.BasicRes) errors.Error {
	return migrationhelper.AutoMigrateTables(
		basicRes,
		&claudeConnection20260526{},
		&claudeWorkspace20260526{},
		&claudeScopeConfig20260526{},
		&claudeCostEntry20260526{},
		&claudeUsageEntry20260526{},
	)
}

func (s *addClaudeInitialTables) Version() uint64 {
	return 20260527000001
}

func (s *addClaudeInitialTables) Name() string {
	return "add claude-code initial tables"
}

type claudeConnection20260526 struct {
	archived.Model
	Name             string `gorm:"type:varchar(100);uniqueIndex"`
	Endpoint         string `gorm:"type:varchar(255)"`
	Proxy            string `gorm:"type:varchar(255)"`
	RateLimitPerHour int
	Token            string `gorm:"type:varchar(255)"`
}

func (claudeConnection20260526) TableName() string { return "_tool_claude_connections" }

type claudeWorkspace20260526 struct {
	archived.NoPKModel
	ConnectionId  uint64 `gorm:"primaryKey"`
	ScopeConfigId uint64
	Id            string `gorm:"primaryKey;type:varchar(255)"`
	Name          string `gorm:"type:varchar(255)"`
	DisplayName   string `gorm:"type:varchar(255)"`
}

func (claudeWorkspace20260526) TableName() string { return "_tool_claude_workspaces" }

type claudeScopeConfig20260526 struct {
	archived.Model
	ConnectionId uint64 `gorm:"index"`
	Entities     string `gorm:"type:json"`
	Name         string `gorm:"type:varchar(255);uniqueIndex"`
}

func (claudeScopeConfig20260526) TableName() string { return "_tool_claude_scope_configs" }

type claudeCostEntry20260526 struct {
	archived.NoPKModel
	ConnectionId  uint64    `gorm:"primaryKey"`
	ScopeId       string    `gorm:"primaryKey;type:varchar(255)"`
	StartingAt    time.Time `gorm:"primaryKey"`
	EndingAt      time.Time
	WorkspaceId   string  `gorm:"type:varchar(255)"`
	Model         string  `gorm:"type:varchar(255)"`
	TokenType     string  `gorm:"type:varchar(100)"`
	CostType      string  `gorm:"type:varchar(100)"`
	ContextWindow string  `gorm:"type:varchar(50)"`
	ServiceTier   string  `gorm:"type:varchar(50)"`
	InferenceGeo  string  `gorm:"type:varchar(100)"`
	Amount        float64
	Currency      string `gorm:"type:varchar(10)"`
	Description   string `gorm:"type:text"`
}

func (claudeCostEntry20260526) TableName() string { return "_tool_claude_cost_entries" }

type claudeUsageEntry20260526 struct {
	archived.NoPKModel
	ConnectionId                        uint64    `gorm:"primaryKey"`
	ScopeId                             string    `gorm:"primaryKey;type:varchar(255)"`
	StartingAt                          time.Time `gorm:"primaryKey"`
	EndingAt                            time.Time
	WorkspaceId                         string `gorm:"type:varchar(255)"`
	Model                               string `gorm:"type:varchar(255)"`
	AccountId                           string `gorm:"type:varchar(255)"`
	ApiKeyId                            string `gorm:"type:varchar(255)"`
	ServiceAccountId                    string `gorm:"type:varchar(255)"`
	ServiceTier                         string `gorm:"type:varchar(50)"`
	UncachedInputTokens                 int64
	OutputTokens                        int64
	CacheReadInputTokens                int64
	CacheCreationEphemeral1hInputTokens int64
	CacheCreationEphemeral5mInputTokens int64
}

func (claudeUsageEntry20260526) TableName() string { return "_tool_claude_usage_entries" }
