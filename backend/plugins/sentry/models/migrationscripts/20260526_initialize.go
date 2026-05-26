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
)

type addSentryInitialTables struct{}

func (s *addSentryInitialTables) Up(basicRes context.BasicRes) errors.Error {
	return migrationhelper.AutoMigrateTables(
		basicRes,
		&sentryConnection20260526{},
		&sentryProject20260526{},
		&sentryScopeConfig20260526{},
		&sentryIssue20260526{},
	)
}

func (s *addSentryInitialTables) Version() uint64 {
	return 20260526000001
}

func (s *addSentryInitialTables) Name() string {
	return "add sentry initial tables"
}

// sentryConnection20260526 is the initial connection schema.
type sentryConnection20260526 struct {
	archived.Model
	Name             string `gorm:"type:varchar(100);uniqueIndex"`
	Endpoint         string `gorm:"type:varchar(255)"`
	Proxy            string `gorm:"type:varchar(255)"`
	RateLimitPerHour int
	Token            string `gorm:"type:varchar(255)"`
}

func (sentryConnection20260526) TableName() string { return "_tool_sentry_connections" }

// sentryProject20260526 is the initial project/scope schema.
type sentryProject20260526 struct {
	archived.NoPKModel
	ConnectionId uint64 `gorm:"primaryKey"`
	ScopeConfigId uint64
	Id           string `gorm:"primaryKey;type:varchar(255)"`
	OrgSlug      string `gorm:"type:varchar(255)"`
	Slug         string `gorm:"type:varchar(255)"`
	Name         string `gorm:"type:varchar(255)"`
	Platform     string `gorm:"type:varchar(100)"`
}

func (sentryProject20260526) TableName() string { return "_tool_sentry_projects" }

// sentryScopeConfig20260526 is the initial scope config schema.
type sentryScopeConfig20260526 struct {
	archived.Model
	ConnectionId uint64 `gorm:"primaryKey"`
	Name         string `gorm:"type:varchar(255)"`
}

func (sentryScopeConfig20260526) TableName() string { return "_tool_sentry_scope_configs" }

// sentryIssue20260526 is the initial issue schema.
type sentryIssue20260526 struct {
	archived.NoPKModel
	ConnectionId  uint64 `gorm:"primaryKey"`
	ScopeId       string `gorm:"primaryKey;type:varchar(255)"`
	IssueId       string `gorm:"primaryKey;type:varchar(255)"`
	ShortId       string `gorm:"type:varchar(255)"`
	Title         string `gorm:"type:text"`
	Culprit       string `gorm:"type:text"`
	Permalink     string `gorm:"type:varchar(500)"`
	Level         string `gorm:"type:varchar(50)"`
	Status        string `gorm:"type:varchar(50)"`
	Platform      string `gorm:"type:varchar(100)"`
	IssueType     string `gorm:"type:varchar(100)"`
	IssueCategory string `gorm:"type:varchar(100)"`
	Priority      string `gorm:"type:varchar(50)"`
	AssigneeName  string `gorm:"type:varchar(255)"`
	AssigneeEmail string `gorm:"type:varchar(255)"`
	Count         string `gorm:"type:varchar(50)"`
	UserCount     int
	FirstSeen     *archived.Iso8601Time
	LastSeen      *archived.Iso8601Time
	OrgSlug       string `gorm:"type:varchar(255)"`
	ProjectSlug   string `gorm:"type:varchar(255)"`
}

func (sentryIssue20260526) TableName() string { return "_tool_sentry_issues" }
