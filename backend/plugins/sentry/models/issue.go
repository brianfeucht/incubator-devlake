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

// SentryIssue is the tool-layer model for a Sentry Issue (group).
type SentryIssue struct {
	common.NoPKModel
	ConnectionId  uint64     `json:"connectionId" gorm:"primaryKey"`
	ScopeId       string     `json:"scopeId" gorm:"primaryKey;type:varchar(255)"`
	IssueId       string     `json:"issueId" gorm:"primaryKey;type:varchar(255)"`
	ShortId       string     `json:"shortId" gorm:"type:varchar(255)"`
	Title         string     `json:"title" gorm:"type:text"`
	Culprit       string     `json:"culprit" gorm:"type:text"`
	Permalink     string     `json:"permalink" gorm:"type:varchar(500)"`
	Level         string     `json:"level" gorm:"type:varchar(50)"`
	Status        string     `json:"status" gorm:"type:varchar(50)"`
	Platform      string     `json:"platform" gorm:"type:varchar(100)"`
	IssueType     string     `json:"issueType" gorm:"type:varchar(100)"`
	IssueCategory string     `json:"issueCategory" gorm:"type:varchar(100)"`
	Priority      string     `json:"priority" gorm:"type:varchar(50)"`
	AssigneeName  string     `json:"assigneeName" gorm:"type:varchar(255)"`
	AssigneeEmail string     `json:"assigneeEmail" gorm:"type:varchar(255)"`
	Count         string     `json:"count" gorm:"type:varchar(50)"`
	UserCount     int        `json:"userCount"`
	FirstSeen     *time.Time `json:"firstSeen"`
	LastSeen      *time.Time `json:"lastSeen"`
	// OrgSlug and ProjectSlug for reference
	OrgSlug       string `json:"orgSlug" gorm:"type:varchar(255)"`
	ProjectSlug   string `json:"projectSlug" gorm:"type:varchar(255)"`
}

func (SentryIssue) TableName() string {
	return "_tool_sentry_issues"
}
