// Copyright (c) 2023 Red Hat, Inc.

package version

// This file hosts a k8s suitable version object.

import "k8s.io/apimachinery/pkg/version"

var (
	tag    = "replace-me"
	commit = "replace-me"
	date   = "replace-me"
	gover  = "replace-me"
)

// Get is used for retrieving this project's version.Info.
func Get() version.Info {
	return version.Info{
		GitVersion: tag,
		GitCommit:  commit,
		BuildDate:  date,
		GoVersion:  gover,
	}
}
