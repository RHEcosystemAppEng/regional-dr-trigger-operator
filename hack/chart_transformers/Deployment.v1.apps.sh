#!/bin/bash

# Copyright (c) 2023 Red Hat, Inc.

# iterate over arguments and create named parameters
while [ $# -gt 0 ]; do
	if [[ $1 == *"--"* ]]; then
		param="${1/--/}"
		declare "$param"="$2"
	fi
	shift
done

# mandatory named parameters
[[ -z $target_manifest ]] && echo "missing mandatory target_manifest" && exit 1

temp_folder=hack/chart_tmp

# set namespace template
yq -i '.metadata.namespace = "{{ .Values.operator.namespace }}"' "$target_manifest"

# fetch current replicas
replicas=$(yq '.spec.replicas' "$target_manifest")
# set current replicas in values
yq -i ".operator.replicas = $replicas" "$temp_folder"/values.yaml
# set replicas template
yq -i '.spec.replicas = "{{ .Values.operator.replicas | int }}"' "$target_manifest"

# iterate over containers, here we go over fields we want to replace with a template in each container, move them to
# the values.yaml file, and replace them with a suitable tempalte
containers=$(yq '.spec.template.spec.containers[] | .name' "$target_manifest")
for container in $containers; do
    container_fixed="${container//-/_}"

    # handle image
    image=$(yq ".spec.template.spec.containers[] | select(.name == \"$container\") | .image" "$target_manifest")
    yq -i ".operator.$container_fixed.image = \"$image\"" "$temp_folder"/values.yaml
    yq -i "(.spec.template.spec.containers[] | select(.name == \"$container\") | .image) = \"{{ .Values.operator.$container_fixed.image }}\"" "$target_manifest"

    # handle imagePullPolicy
    image_pull_policy=$(yq ".spec.template.spec.containers[] | select(.name == \"$container\") | .imagePullPolicy" "$target_manifest")
    yq -i ".operator.$container_fixed.imagePullPolicy = \"$image_pull_policy\"" "$temp_folder"/values.yaml
    yq -i "(.spec.template.spec.containers[] | select(.name == \"$container\") | .imagePullPolicy) = \"{{ .Values.operator.$container_fixed.imagePullPolicy }}\"" "$target_manifest"

    # handle resources
    resources=$(yq -o json ".spec.template.spec.containers[] | select(.name == \"$container\") | .resources" "$target_manifest")
    yq -i ".operator.$container_fixed.resources = $resources" "$temp_folder"/values.yaml
    yq -i "(.spec.template.spec.containers[] | select(.name == \"$container\") | .resources) = \"{{ .Values.operator.$container_fixed.resources | toYaml | nindent 12 }}\"" "$target_manifest"
done
