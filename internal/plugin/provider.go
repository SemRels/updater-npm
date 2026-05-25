// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The updater-npm Authors

package plugin

import (
	"context"
	"errors"
	"os"

	semrelv1 "github.com/SemRels/updater-npm/internal/gen/v1"
	grpcserver "github.com/SemRels/updater-npm/internal/grpc"
	hashiplugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

const (
	FilesUpdaterPluginName  = "files-updater"
	DefaultProtocolVersion  = 1
	defaultMagicCookieKey   = "SEMREL_PLUGIN"
	defaultMagicCookieValue = "updater-npm"
)

const (
	magicCookieKeyEnv   = "SEMREL_PLUGIN_MAGIC_COOKIE_KEY"
	magicCookieValueEnv = "SEMREL_PLUGIN_MAGIC_COOKIE_VALUE"
)

// FilesUpdater defines the business contract exposed through the plugin transport.
type FilesUpdater interface {
	UpdateFiles(context.Context, *semrelv1.ReleaseContext) ([]string, error)
}

// GRPCPlugin exposes the files updater over hashicorp/go-plugin gRPC transport.
type GRPCPlugin struct {
	hashiplugin.Plugin
	Impl FilesUpdater
}

func (p *GRPCPlugin) GRPCServer(_ *hashiplugin.GRPCBroker, server *grpc.Server) error {
	semrelv1.RegisterFilesUpdaterPluginServer(server, grpcserver.NewFilesUpdaterServer(p.Impl))
	return nil
}

func (p *GRPCPlugin) GRPCClient(_ context.Context, _ *hashiplugin.GRPCBroker, conn *grpc.ClientConn) (interface{}, error) {
	return &Client{client: semrelv1.NewFilesUpdaterPluginClient(conn)}, nil
}

// Client adapts the generated gRPC client to the FilesUpdater interface.
type Client struct {
	client semrelv1.FilesUpdaterPluginClient
}

func (c *Client) UpdateFiles(ctx context.Context, releaseContext *semrelv1.ReleaseContext) ([]string, error) {
	response, err := c.client.UpdateFiles(ctx, &semrelv1.UpdateFilesRequest{Ctx: releaseContext})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, errors.New("empty UpdateFiles response")
	}
	if response.GetErrorMessage() != "" {
		return response.GetUpdatedFiles(), errors.New(response.GetErrorMessage())
	}
	return response.GetUpdatedFiles(), nil
}

func PluginMap(updater FilesUpdater) map[string]hashiplugin.Plugin {
	return map[string]hashiplugin.Plugin{
		FilesUpdaterPluginName: &GRPCPlugin{Impl: updater},
	}
}

func HandshakeConfig() hashiplugin.HandshakeConfig {
	return hashiplugin.HandshakeConfig{
		ProtocolVersion:  DefaultProtocolVersion,
		MagicCookieKey:   envOrDefault(magicCookieKeyEnv, defaultMagicCookieKey),
		MagicCookieValue: envOrDefault(magicCookieValueEnv, defaultMagicCookieValue),
	}
}

func envOrDefault(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
