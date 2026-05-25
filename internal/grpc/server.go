// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The updater-npm Authors

package grpc

import (
	"context"

	semrelv1 "github.com/SemRels/updater-npm/internal/gen/v1"
)

// Updater is the minimal contract required by the gRPC adapter.
type Updater interface {
	UpdateFiles(context.Context, *semrelv1.ReleaseContext) ([]string, error)
}

// FilesUpdaterServer adapts the updater implementation to the protobuf service.
type FilesUpdaterServer struct {
	semrelv1.UnimplementedFilesUpdaterPluginServer
	updater Updater
}

func NewFilesUpdaterServer(updater Updater) *FilesUpdaterServer {
	return &FilesUpdaterServer{updater: updater}
}

func (s *FilesUpdaterServer) UpdateFiles(ctx context.Context, request *semrelv1.UpdateFilesRequest) (*semrelv1.UpdateFilesResponse, error) {
	response := &semrelv1.UpdateFilesResponse{}
	if s.updater == nil {
		response.ErrorMessage = "updater is not configured"
		return response, nil
	}

	updatedFiles, err := s.updater.UpdateFiles(ctx, request.GetCtx())
	response.UpdatedFiles = updatedFiles
	if err != nil {
		response.ErrorMessage = err.Error()
	}
	return response, nil
}
