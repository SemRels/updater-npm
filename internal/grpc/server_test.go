// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The updater-npm Authors

package grpc

import (
	"context"
	"errors"
	"testing"

	semrelv1 "github.com/SemRels/updater-npm/internal/gen/v1"
	"github.com/stretchr/testify/require"
)

func TestFilesUpdaterServerReturnsUpdatedFiles(t *testing.T) {
	t.Parallel()

	server := NewFilesUpdaterServer(stubUpdater{updatedFiles: []string{"package.json"}})

	response, err := server.UpdateFiles(context.Background(), &semrelv1.UpdateFilesRequest{Ctx: &semrelv1.ReleaseContext{}})
	require.NoError(t, err)
	require.Equal(t, []string{"package.json"}, response.GetUpdatedFiles())
	require.Empty(t, response.GetErrorMessage())
}

func TestFilesUpdaterServerReturnsErrorMessage(t *testing.T) {
	t.Parallel()

	server := NewFilesUpdaterServer(stubUpdater{err: errors.New("invalid package.json")})

	response, err := server.UpdateFiles(context.Background(), &semrelv1.UpdateFilesRequest{Ctx: &semrelv1.ReleaseContext{}})
	require.NoError(t, err)
	require.Equal(t, "invalid package.json", response.GetErrorMessage())
}

type stubUpdater struct {
	updatedFiles []string
	err          error
}

func (s stubUpdater) UpdateFiles(context.Context, *semrelv1.ReleaseContext) ([]string, error) {
	return s.updatedFiles, s.err
}
