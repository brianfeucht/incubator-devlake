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
	"gorm.io/gorm"
)

// SentryProject represents a Sentry project used as a DevLake data scope.
// The scope ID is "<org_slug>/<project_slug>".
type SentryProject struct {
	common.Scope  `mapstructure:",squash"`
	Id            string `json:"id" mapstructure:"id" gorm:"primaryKey;type:varchar(255)"`
	OrgSlug       string `json:"orgSlug" mapstructure:"orgSlug" gorm:"type:varchar(255)"`
	Slug          string `json:"slug" mapstructure:"slug" gorm:"type:varchar(255)"`
	Name          string `json:"name" mapstructure:"name" gorm:"type:varchar(255)"`
	Platform      string `json:"platform" mapstructure:"platform" gorm:"type:varchar(100)"`
}

func (SentryProject) TableName() string {
	return "_tool_sentry_projects"
}

func (s SentryProject) ScopeId() string {
	return s.Id
}

func (s SentryProject) ScopeName() string {
	if s.Name != "" {
		return s.Name
	}
	return s.Id
}

func (s SentryProject) ScopeFullName() string {
	return s.Id
}

func (s SentryProject) ScopeParams() interface{} {
	return &SentryProjectParams{
		ConnectionId: s.ConnectionId,
		ScopeId:      s.Id,
	}
}

type SentryProjectParams struct {
	ConnectionId uint64 `json:"connectionId"`
	ScopeId      string `json:"scopeId"`
}

var _ plugin.ToolLayerScope = (*SentryProject)(nil)

// BeforeSave ensures the Name field is set before writing to the database.
func (s *SentryProject) BeforeSave(_ *gorm.DB) error {
	if s.Name == "" {
		s.Name = s.Slug
	}
	return nil
}
