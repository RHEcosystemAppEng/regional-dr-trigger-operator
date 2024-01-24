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

# optional named parameters default values
temp_folder=${temp_folder:-hack/chart_tmp}
bin_yq=${bin_yq:-yq}

# mandatory named parameters
[[ -z $target_manifest ]] && echo "missing mandatory target_manifest" && exit 1

# set namespace template
$bin_yq -i '.metadata.namespace = "{{ .Values.operator.namespace }}"' "$target_manifest"

# fetch current replicas
replicas=$($bin_yq '.spec.replicas' "$target_manifest")
# set current replicas in values
$bin_yq -i ".operator.replicas = $replicas" "$temp_folder"/values.yaml
# set replicas template
$bin_yq -i '.spec.replicas = "{{ .Values.operator.replicas | int }}"' "$target_manifest"

# iterate over containers, here we go over fields we want to replace with a template in each container, move them to
# the values.yaml file, and replace them with a suitable tempalte
containers=$($bin_yq '.spec.template.spec.containers[] | .name' "$target_manifest")
for container in $containers; do
    container_fixed="${container//-/_}"

    # handle image
    image=$($bin_yq ".spec.template.spec.containers[] | select(.name == \"$container\") | .image" "$target_manifest")
    $bin_yq -i ".operator.$container_fixed.image = \"$image\"" "$temp_folder"/values.yaml
    $bin_yq -i "(.spec.template.spec.containers[] | select(.name == \"$container\") | .image) = \"{{ .Values.operator.$container_fixed.image }}\"" "$target_manifest"

    # handle imagePullPolicy
    image_pull_policy=$($bin_yq ".spec.template.spec.containers[] | select(.name == \"$container\") | .imagePullPolicy" "$target_manifest")
    $bin_yq -i ".operator.$container_fixed.imagePullPolicy = \"$image_pull_policy\"" "$temp_folder"/values.yaml
    $bin_yq -i "(.spec.template.spec.containers[] | select(.name == \"$container\") | .imagePullPolicy) = \"{{ .Values.operator.$container_fixed.imagePullPolicy }}\"" "$target_manifest"

    # handle resources
    resources=$($bin_yq -o json ".spec.template.spec.containers[] | select(.name == \"$container\") | .resources" "$target_manifest")
    $bin_yq -i ".operator.$container_fixed.resources = $resources" "$temp_folder"/values.yaml
    $bin_yq -i "(.spec.template.spec.containers[] | select(.name == \"$container\") | .resources) = \"{{ .Values.operator.$container_fixed.resources | toYaml | nindent 12 }}\"" "$target_manifest"
done
