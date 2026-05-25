// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The semrel Authors

package main

import (
	"log"

	plugin "github.com/SemRels/updater-npm/internal/plugin"
)

func main() {
	publisher := plugin.NewPublisher(plugin.Config{})
	log.Printf("updater-npm plugin ready: updates package.json and publishes npm packages (%T)", publisher)
}
