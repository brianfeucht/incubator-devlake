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

type addLinearInitialTables struct{}

func (s *addLinearInitialTables) Up(basicRes context.BasicRes) errors.Error {
	return migrationhelper.AutoMigrateTables(
		basicRes,
		&linearConnection20260527{},
		&linearTeam20260527{},
		&linearScopeConfig20260527{},
		&linearIssue20260527{},
		&linearCycle20260527{},
	)
}

func (s *addLinearInitialTables) Version() uint64 { return 20260527000002 }
func (s *addLinearInitialTables) Name() string    { return "add linear initial tables" }

type linearConnection20260527 struct {
	archived.Model
	Name             string `gorm:"type:varchar(100);uniqueIndex"`
	Endpoint         string `gorm:"type:varchar(255)"`
	Proxy            string `gorm:"type:varchar(255)"`
	RateLimitPerHour int
	Token            string `gorm:"type:varchar(255)"`
}

func (linearConnection20260527) TableName() string { return "_tool_linear_connections" }

type linearTeam20260527 struct {
	archived.NoPKModel
	ConnectionId  uint64 `gorm:"primaryKey"`
	ScopeConfigId uint64
	Id            string `gorm:"primaryKey;type:varchar(255)"`
	Name          string `gorm:"type:varchar(255)"`
	Key           string `gorm:"type:varchar(50)"`
	Description   string `gorm:"type:varchar(1000)"`
}

func (linearTeam20260527) TableName() string { return "_tool_linear_teams" }

type linearScopeConfig20260527 struct {
	archived.Model
	ConnectionId uint64 `gorm:"primaryKey"`
	Name         string `gorm:"type:varchar(255)"`
}

func (linearScopeConfig20260527) TableName() string { return "_tool_linear_scope_configs" }

type linearIssue20260527 struct {
	archived.NoPKModel
	ConnectionId  uint64  `gorm:"primaryKey"`
	TeamId        string  `gorm:"primaryKey;type:varchar(255)"`
	Id            string  `gorm:"primaryKey;type:varchar(255)"`
	Number        int
	Title         string  `gorm:"type:varchar(1000)"`
	Description   string  `gorm:"type:text"`
	Priority      int
	PriorityLabel string  `gorm:"type:varchar(50)"`
	Estimate      *float64
	StateId       string  `gorm:"type:varchar(255)"`
	StateName     string  `gorm:"type:varchar(255)"`
	StateType     string  `gorm:"type:varchar(50)"`
	AssigneeId    string  `gorm:"type:varchar(255)"`
	AssigneeName  string  `gorm:"type:varchar(255)"`
	AssigneeEmail string  `gorm:"type:varchar(255)"`
	CreatorId     string  `gorm:"type:varchar(255)"`
	CreatorName   string  `gorm:"type:varchar(255)"`
	CycleId       string  `gorm:"type:varchar(255)"`
	Url           string  `gorm:"type:varchar(2048)"`
}

func (linearIssue20260527) TableName() string { return "_tool_linear_issues" }

type linearCycle20260527 struct {
	archived.NoPKModel
	ConnectionId uint64 `gorm:"primaryKey"`
	TeamId       string `gorm:"primaryKey;type:varchar(255)"`
	Id           string `gorm:"primaryKey;type:varchar(255)"`
	Number       int
	Name         string `gorm:"type:varchar(255)"`
}

func (linearCycle20260527) TableName() string { return "_tool_linear_cycles" }
