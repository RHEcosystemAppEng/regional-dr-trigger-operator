// Copyright (c) 2023 Red Hat, Inc.

package version

import "k8s.io/apimachinery/pkg/version"

func Get() version.Info {
	return version.Info{
		GitVersion: "TODO-Replace-Me-And-Add-Fields",
	}
}
