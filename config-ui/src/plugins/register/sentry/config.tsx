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

export const SentryConfig: IPluginConfig = {
  plugin: 'sentry',
  name: 'Sentry',
  icon: ({ color }) => <Icon fill={color} />,
  sort: 18.5,
  isBeta: true,
  connection: {
    docLink: 'https://docs.sentry.io/api/',
    initialValues: {
      endpoint: 'https://sentry.io',
      token: '',
      rateLimitPerHour: 3600,
    },
    fields: [
      'name',
      'endpoint',
      {
        key: 'token',
        label: 'Auth Token',
        subLabel:
          'Create a User Auth Token at https://sentry.io/settings/account/api/auth-tokens/ with scopes: project:read, event:read, org:read.',
      },
      'proxy',
      {
        key: 'rateLimitPerHour',
        subLabel: 'By default, DevLake uses 3,600 requests/hour for Sentry data collection.',
        defaultValue: 3600,
      },
    ],
  },
  dataScope: {
    title: 'Projects',
    millerColumn: {
      columnCount: 2,
      firstColumnTitle: 'Organizations',
    },
  },
  scopeConfig: {
    entities: ['TICKET'],
    transformation: {},
  },
};
