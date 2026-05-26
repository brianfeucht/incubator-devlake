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

// ClaudeCostEntry holds one cost line-item returned by the Anthropic cost_report API.
// Each API response bucket contains multiple results, one per model/token_type combo.
type ClaudeCostEntry struct {
	common.NoPKModel
	ConnectionId  uint64    `gorm:"primaryKey"`
	ScopeId       string    `gorm:"primaryKey;type:varchar(255)"` // workspace ID
	StartingAt    time.Time `gorm:"primaryKey"`
	EndingAt      time.Time
	WorkspaceId   string  `gorm:"type:varchar(255)"`
	Model         string  `gorm:"type:varchar(255)"`
	TokenType     string  `gorm:"type:varchar(100)"` // uncached_input_tokens, output_tokens, etc.
	CostType      string  `gorm:"type:varchar(100)"` // tokens, web_search, code_execution, session_usage
	ContextWindow string  `gorm:"type:varchar(50)"`  // 0-200k, 200k-1M, etc.
	ServiceTier   string  `gorm:"type:varchar(50)"`  // standard, batch
	InferenceGeo  string  `gorm:"type:varchar(100)"`
	Amount        float64 // parsed from string
	Currency      string  `gorm:"type:varchar(10)"`
	Description   string  `gorm:"type:text"`
}

func (ClaudeCostEntry) TableName() string {
	return "_tool_claude_cost_entries"
}
