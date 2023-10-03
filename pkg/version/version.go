// Copyright (c) 2023 Red Hat, Inc.

package version

import "k8s.io/apimachinery/pkg/version"

// verInfo is the version-info for this project
var verInfo = version.Info{
	GitVersion: "TODO-Replace-Me-And-Add-Fields",
}

// Get is used for retrieving this project's version info.
func Get() version.Info {
	return verInfo
}
