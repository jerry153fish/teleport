/**
 * Copyright 2021 Gravitational, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { useState } from 'react';
import * as Alerts from 'design/Alert';
import { Box, Text, ButtonPrimary, ButtonSecondary } from 'design';
import FieldInput from 'shared/components/FieldInput';
import Validation from 'shared/components/Validation';
import { requiredField } from 'shared/components/Validation/rules';
import { DialogContent, DialogHeader } from 'design/Dialog';
import { useAsync } from 'shared/hooks/useAsync';

import { useAppContext } from 'teleterm/ui/appContextProvider';

export function ClusterAdd(props: {
  onCancel(): void;
  onSuccess(clusterUri: string): void;
  prefill: { clusterAddress: string };
}) {
  const { clustersService } = useAppContext();
  const [{ status, statusText }, addCluster] = useAsync(
    async (addr: string) => {
      const proxyAddr = parseClusterProxyWebAddr(addr);
      const cluster = await clustersService.addRootCluster(proxyAddr);
      return props.onSuccess(cluster.uri);
    }
  );
  const [addr, setAddr] = useState(props.prefill.clusterAddress || '');

  return (
    <Box p={4}>
      <Validation>
        {({ validator }) => (
          <form
            onSubmit={e => {
              e.preventDefault();
              validator.validate() && addCluster(addr);
            }}
          >
            <DialogHeader>
              <Text typography="h4">Enter cluster address</Text>
            </DialogHeader>
            <DialogContent mb={2}>
              {status === 'error' && (
                <Alerts.Danger mb={5} children={statusText} />
              )}
              <FieldInput
                rule={requiredField('Cluster address is required')}
                value={addr}
                autoFocus
                onChange={e => setAddr(e.target.value)}
                placeholder="teleport.example.com"
              />
              <Box mt="5">
                <ButtonPrimary
                  disabled={status === 'processing'}
                  mr="3"
                  type="submit"
                >
                  Next
                </ButtonPrimary>
                <ButtonSecondary
                  disabled={status === 'processing'}
                  type="button"
                  onClick={e => {
                    e.preventDefault();
                    props.onCancel();
                  }}
                >
                  CANCEL
                </ButtonSecondary>
              </Box>
            </DialogContent>
          </form>
        )}
      </Validation>
    </Box>
  );
}

function parseClusterProxyWebAddr(addr: string) {
  addr = addr || '';
  if (addr.startsWith('http')) {
    const url = new URL(addr);
    return url.host;
  }

  return addr;
}
