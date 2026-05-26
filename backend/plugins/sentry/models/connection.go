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
	"net/http"
	"strings"

	"github.com/apache/incubator-devlake/core/errors"
	"github.com/apache/incubator-devlake/core/utils"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
)

const (
	DefaultEndpoint        = "https://sentry.io"
	DefaultRateLimitPerHour = 3600
)

// SentryConn holds Sentry connection settings.
type SentryConn struct {
	helper.RestConnection `mapstructure:",squash"`
	Token                 string `mapstructure:"token" json:"token"`
	RateLimitPerHour      int    `mapstructure:"rateLimitPerHour" json:"rateLimitPerHour"`
}

// SetupAuthentication attaches the Bearer token to every API request.
func (conn *SentryConn) SetupAuthentication(req *http.Request) errors.Error {
	if conn == nil || strings.TrimSpace(conn.Token) == "" {
		return errors.BadInput.New("token is required")
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(conn.Token))
	return nil
}

func (conn *SentryConn) Sanitize() SentryConn {
	clone := *conn
	clone.Token = utils.SanitizeString(clone.Token)
	return clone
}

// SentryConnection persists connection details in the database.
type SentryConnection struct {
	helper.BaseConnection `mapstructure:",squash"`
	SentryConn            `mapstructure:",squash"`
}

func (SentryConnection) TableName() string {
	return "_tool_sentry_connections"
}

func (c SentryConnection) Sanitize() SentryConnection {
	c.SentryConn = c.SentryConn.Sanitize()
	return c
}

func (c *SentryConnection) MergeFromRequest(target *SentryConnection, body map[string]interface{}) error {
	originalToken := target.Token
	if err := helper.DecodeMapStruct(body, target, true); err != nil {
		return err
	}
	if target.Token == "" || target.Token == utils.SanitizeString(originalToken) {
		target.Token = originalToken
	}
	return nil
}

// Normalize fills default values.
func (c *SentryConnection) Normalize() {
	if c.Endpoint == "" {
		c.Endpoint = DefaultEndpoint
	}
	if c.RateLimitPerHour <= 0 {
		c.RateLimitPerHour = DefaultRateLimitPerHour
	}
}
