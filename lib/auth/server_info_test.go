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

package auth

import (
	"context"
	"testing"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/api/client"
	"github.com/gravitational/teleport/api/client/proto"
	"github.com/gravitational/teleport/api/types"
)

type mockUpstream struct {
	client.UpstreamInventoryControlStream
	updatedLabels map[string]string
}

func (m *mockUpstream) Send(_ context.Context, msg proto.DownstreamInventoryMessage) error {
	if labelMsg, ok := msg.(proto.DownstreamInventoryUpdateLabels); ok {
		m.updatedLabels = labelMsg.Labels
	}
	return nil
}

func (m *mockUpstream) Recv() <-chan proto.UpstreamInventoryMessage {
	return make(chan proto.UpstreamInventoryMessage)
}

func (m *mockUpstream) Done() <-chan struct{} {
	return make(chan struct{})
}

func (m *mockUpstream) Close() error {
	return nil
}

// TestReconcileLabels verifies that an SSH server's labels can be updated by
// upserting a corresponding ServerInfo to the auth server.
func TestReconcileLabels(t *testing.T) {
	t.Parallel()

	const serverName = "test-server"
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Create auth server and fake inventory stream.
	clock := clockwork.NewFakeClock()
	pack, err := newTestPack(ctx, t.TempDir(), WithClock(clock))
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, pack.a.Close())
		require.NoError(t, pack.bk.Close())
	})
	upstream := &mockUpstream{}
	t.Cleanup(func() {
		require.NoError(t, upstream.Close())
	})
	require.NoError(t, pack.a.RegisterInventoryControlStream(upstream, proto.UpstreamInventoryHello{
		Version:  teleport.Version,
		ServerID: serverName,
		Services: []types.SystemRole{types.RoleNode},
	}))

	// Create server.
	server, err := types.NewServer(serverName, types.KindNode, types.ServerSpecV2{
		CloudMetadata: &types.CloudMetadata{
			AWS: &types.AWSInfo{
				AccountID:  "my-account",
				InstanceID: "my-instance",
			},
		},
	})
	require.NoError(t, err)
	_, err = pack.a.UpsertNode(ctx, server)
	require.NoError(t, err)

	// Update the server's labels.
	awsServerInfo, err := types.NewServerInfo(types.Metadata{
		Name: types.ServerInfoNameFromAWS("my-account", "my-instance"),
	}, types.ServerInfoSpecV1{
		NewLabels: map[string]string{"a": "1", "b": "2"},
	})
	require.NoError(t, err)
	require.NoError(t, pack.a.UpsertServerInfo(ctx, awsServerInfo))

	regularServerInfo, err := types.NewServerInfo(types.Metadata{
		Name: types.ServerInfoNameFromNodeName(serverName),
	}, types.ServerInfoSpecV1{
		NewLabels: map[string]string{"b": "3", "c": "4"},
	})
	require.NoError(t, err)
	require.NoError(t, pack.a.UpsertServerInfo(ctx, regularServerInfo))

	go pack.a.ReconcileServerInfos(ctx)
	// Wait until the reconciler finishes processing the serverinfo.
	clock.BlockUntil(1)
	// Check that labels were received downstream.
	require.Equal(t, map[string]string{"a": "1", "b": "3", "c": "4"}, upstream.updatedLabels)
}
