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
	"time"

	"github.com/apache/incubator-devlake/core/context"
	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/helpers/migrationhelper"
)

type addLinearIssueDateColumns struct{}

func (s *addLinearIssueDateColumns) Up(basicRes context.BasicRes) errors.Error {
	return migrationhelper.AutoMigrateTables(basicRes, &linearIssue20260527b{})
}

func (s *addLinearIssueDateColumns) Version() uint64 { return 20260527000003 }
func (s *addLinearIssueDateColumns) Name() string    { return "add date columns to linear issues" }

// linearIssue20260527b is the full schema including all date/time columns.
type linearIssue20260527b struct {
	ConnectionId  uint64     `gorm:"primaryKey"`
	TeamId        string     `gorm:"primaryKey;type:varchar(255)"`
	Id            string     `gorm:"primaryKey;type:varchar(255)"`
	Number        int
	Title         string     `gorm:"type:varchar(1000)"`
	Description   string     `gorm:"type:text"`
	Priority      int
	PriorityLabel string     `gorm:"type:varchar(50)"`
	Estimate      *float64
	StateId       string     `gorm:"type:varchar(255)"`
	StateName     string     `gorm:"type:varchar(255)"`
	StateType     string     `gorm:"type:varchar(50)"`
	AssigneeId    string     `gorm:"type:varchar(255)"`
	AssigneeName  string     `gorm:"type:varchar(255)"`
	AssigneeEmail string     `gorm:"type:varchar(255)"`
	CreatorId     string     `gorm:"type:varchar(255)"`
	CreatorName   string     `gorm:"type:varchar(255)"`
	CycleId       string     `gorm:"type:varchar(255)"`
	Url           string     `gorm:"type:varchar(2048)"`
	DueDate       *time.Time
	StartedAt     *time.Time
	CompletedAt   *time.Time
	CanceledAt    *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (linearIssue20260527b) TableName() string { return "_tool_linear_issues" }
