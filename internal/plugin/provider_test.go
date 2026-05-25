// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The updater-npm Authors

package plugin

import (
	"context"
	"errors"
	"testing"

	semrelv1 "github.com/SemRels/updater-npm/internal/gen/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestHandshakeConfigDefaults(t *testing.T) {
	handshake := HandshakeConfig()

	require.Equal(t, uint(DefaultProtocolVersion), handshake.ProtocolVersion)
	require.Equal(t, defaultMagicCookieKey, handshake.MagicCookieKey)
	require.Equal(t, defaultMagicCookieValue, handshake.MagicCookieValue)
}

func TestHandshakeConfigHonorsEnvironment(t *testing.T) {
	t.Setenv(magicCookieKeyEnv, "CUSTOM_KEY")
	t.Setenv(magicCookieValueEnv, "custom-value")

	handshake := HandshakeConfig()

	require.Equal(t, "CUSTOM_KEY", handshake.MagicCookieKey)
	require.Equal(t, "custom-value", handshake.MagicCookieValue)
}

func TestClientReturnsRemoteErrorMessage(t *testing.T) {
	t.Parallel()

	client := &Client{client: stubFilesUpdaterClient{response: &semrelv1.UpdateFilesResponse{ErrorMessage: "boom"}}}

	updatedFiles, err := client.UpdateFiles(context.Background(), &semrelv1.ReleaseContext{})
	require.Error(t, err)
	require.EqualError(t, err, "boom")
	require.Empty(t, updatedFiles)
}

func TestPluginMapContainsFilesUpdaterPlugin(t *testing.T) {
	t.Parallel()

	pluginMap := PluginMap(NewUpdater())

	require.Contains(t, pluginMap, FilesUpdaterPluginName)
}

func TestGRPCPluginCreatesClient(t *testing.T) {
	t.Parallel()

	grpcPlugin := &GRPCPlugin{Impl: NewUpdater()}

	client, err := grpcPlugin.GRPCClient(context.Background(), nil, nil)
	require.NoError(t, err)
	require.IsType(t, &Client{}, client)
}

type stubFilesUpdaterClient struct {
	response *semrelv1.UpdateFilesResponse
	err      error
}

func (s stubFilesUpdaterClient) UpdateFiles(context.Context, *semrelv1.UpdateFilesRequest, ...grpc.CallOption) (*semrelv1.UpdateFilesResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.response, nil
}

var _ semrelv1.FilesUpdaterPluginClient = stubFilesUpdaterClient{}

func TestClientPropagatesTransportError(t *testing.T) {
	t.Parallel()

	client := &Client{client: stubFilesUpdaterClient{err: errors.New("transport down")}}

	updatedFiles, err := client.UpdateFiles(context.Background(), &semrelv1.ReleaseContext{})
	require.Error(t, err)
	require.EqualError(t, err, "transport down")
	require.Nil(t, updatedFiles)
}

func TestClientRejectsEmptyResponse(t *testing.T) {
	t.Parallel()

	client := &Client{client: stubFilesUpdaterClient{}}

	updatedFiles, err := client.UpdateFiles(context.Background(), &semrelv1.ReleaseContext{})
	require.Error(t, err)
	require.EqualError(t, err, "empty UpdateFiles response")
	require.Nil(t, updatedFiles)
}
