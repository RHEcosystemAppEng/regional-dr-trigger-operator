// Copyright (c) 2023 Red Hat, Inc.

package version

// This file hosts a k8s suitable version object.

import "k8s.io/apimachinery/pkg/version"

// these are updated in build-time using LDFlags, note the Makefile
var (
	tag    = "replace-me"
	commit = "replace-me"
	date   = "replace-me"
)

// Get is used for retrieving this project's version.Info.
func Get() version.Info {
	return version.Info{
		GitVersion: tag,
		GitCommit:  commit,
		BuildDate:  date,
	}
}
