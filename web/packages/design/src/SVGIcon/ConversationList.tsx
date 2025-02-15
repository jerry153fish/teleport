/*
Copyright 2023 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import React from 'react';

import { SVGIcon } from './SVGIcon';

import type { SVGIconProps } from './common';

export function ConversationListIcon({ size = 24, fill }: SVGIconProps) {
  return (
    <SVGIcon fill={fill} size={size} viewBox="0 0 24 24">
      <path d="M18 8.016V6H6v2.016h12zm-3.984 6V12H6v2.016h8.016zM6 9v2.016h12V9H6zm14.016-6.984q.797 0 1.383.586t.586 1.383v12q0 .797-.586 1.406T20.016 18H6l-3.984 3.984v-18q0-.797.586-1.383t1.383-.586h16.031z" />
    </SVGIcon>
  );
}
