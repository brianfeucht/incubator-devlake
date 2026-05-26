/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import { IPluginConfig } from '@/types';

import Icon from './assets/icon.svg?react';

export const ClaudeCodeConfig: IPluginConfig = {
  plugin: 'claude-code',
  name: 'Claude Code',
  icon: ({ color }) => <Icon fill={color} />,
  sort: 19,
  isBeta: true,
  connection: {
    docLink: 'https://platform.claude.com/docs/en/api/admin/overview',
    initialValues: {
      endpoint: 'https://api.anthropic.com',
      adminApiKey: '',
      rateLimitPerHour: 1200,
    },
    fields: [
      'name',
      'endpoint',
      {
        key: 'adminApiKey',
        label: 'Admin API Key',
        subLabel:
          'Create an Admin API Key in the Anthropic Console at https://console.anthropic.com. The key must have Admin permissions to access cost and usage reports.',
      },
      'proxy',
      {
        key: 'rateLimitPerHour',
        subLabel: 'By default, DevLake uses 1,200 requests/hour for Anthropic data collection.',
        defaultValue: 1200,
      },
    ],
  },
  dataScope: {
    title: 'Workspaces',
    millerColumn: {
      columnCount: 1,
    },
  },
  scopeConfig: {
    entities: ['TICKET'],
    transformation: {},
  },
};
