// Copyright (c) 2023 Red Hat, Inc.

package manager

// This file hosts functions for loading templates.

import rbacv1 "k8s.io/api/rbac/v1"

// This file hosts functions and types for setting out Addon Manager registration process for requesting Agent Addons.

import (
	"bytes"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"text/template"
)

// loadTemplateFromFile is used for loading a template file from fsys, and execute ig against template values. The
// executed template will be generically converted.
func loadTemplateFromFile[T rbacv1.Role | rbacv1.RoleBinding](file string, values interface{}, target *T) error {
	tmpl, err := template.ParseFS(fsys, file)
	if err != nil {
		return err
	}

	var buff bytes.Buffer
	if err = tmpl.Execute(&buff, values); err != nil {
		return err
	}

	manifest := make(map[string]interface{})
	if err = yaml.Unmarshal(buff.Bytes(), &manifest); err != nil {
		return err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(manifest, target); err != nil {
		return err
	}

	return nil
}
