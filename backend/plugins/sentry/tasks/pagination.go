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

package tasks

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/apache/incubator-devlake/core/errors"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
)

// sentryLinkRe matches the cursor value in Sentry's Link response header.
// Example header value:
//   <https://sentry.io/api/0/.../issues/?cursor=0:100:0&...>; rel="next"; results="true"; cursor="0:100:0"
var sentryLinkRe = regexp.MustCompile(`cursor="([^"]+)"`)

// parseSentryLinkCursor is used as GetNextPageCustomData; it extracts the
// "next" cursor from the Link header so the paginator can continue.
func parseSentryLinkCursor(prevReqData *helper.RequestData, res *http.Response) (interface{}, errors.Error) {
	if res == nil {
		return nil, nil
	}
	linkHeader := res.Header.Get("Link")
	if linkHeader == "" {
		return nil, nil
	}

	// The Link header contains multiple entries separated by ", ".
	// We want the one that has rel="next" and results="true".
	for _, part := range strings.Split(linkHeader, ", ") {
		if !strings.Contains(part, `rel="next"`) {
			continue
		}
		if !strings.Contains(part, `results="true"`) {
			// Next page exists in structure but is empty — stop.
			return nil, nil
		}
		m := sentryLinkRe.FindStringSubmatch(part)
		if len(m) == 2 {
			return m[1], nil
		}
	}
	return nil, nil
}

// GetNextPage adapts parseSentryLinkCursor so it satisfies the
// helper.GetNextPageFunc[string] signature used in some collector wrappers.
func GetNextPage(res *http.Response, args *helper.RequestData) (paging *helper.Pager, err errors.Error) {
	cursor, err2 := parseSentryLinkCursor(args, res)
	if err2 != nil {
		return nil, err2
	}
	if cursor == nil {
		return nil, nil
	}
	return &helper.Pager{
		Page: args.Pager.Page + 1,
		Size: args.Pager.Size,
		Skip: 0,
	}, nil
}
