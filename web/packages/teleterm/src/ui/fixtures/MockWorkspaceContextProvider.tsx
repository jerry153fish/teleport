/**
 * Copyright 2023 Gravitational, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from 'react';

import { WorkspaceContextProvider } from 'teleterm/ui/Documents';
import { RootClusterUri } from 'teleterm/ui/uri';
import { useAppContext } from 'teleterm/ui/appContextProvider';

export const MockWorkspaceContextProvider: React.FC<{
  rootClusterUri?: RootClusterUri;
}> = props => {
  const appContext = useAppContext();

  const rootClusterUri =
    props.rootClusterUri || appContext.workspacesService.getRootClusterUri();

  return (
    <WorkspaceContextProvider
      value={{
        accessRequestsService:
          appContext.workspacesService.getWorkspaceAccessRequestsService(
            rootClusterUri
          ),
        documentsService:
          appContext.workspacesService.getWorkspaceDocumentService(
            rootClusterUri
          ),
        localClusterUri: rootClusterUri,
        rootClusterUri,
      }}
    >
      {props.children}
    </WorkspaceContextProvider>
  );
};
