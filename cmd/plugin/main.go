// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The updater-npm Authors

package main

import (
	"os"

	semrelplugin "github.com/SemRels/updater-npm/internal/plugin"
	hclog "github.com/hashicorp/go-hclog"
	hashiplugin "github.com/hashicorp/go-plugin"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "updater-npm",
		Level:  hclog.Info,
		Output: os.Stderr,
	})

	hashiplugin.Serve(&hashiplugin.ServeConfig{
		HandshakeConfig: semrelplugin.HandshakeConfig(),
		Plugins:         semrelplugin.PluginMap(semrelplugin.NewUpdater()),
		GRPCServer:      hashiplugin.DefaultGRPCServer,
		Logger:          logger,
	})
}
