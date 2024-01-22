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

# we have two cluster roles, one for the operator service account, and one for prometheus service account
manifest_name=$($bin_yq '.metadata.name' "$target_manifest")

# for our operator cluster role, set the subject namespace
[[ "regional-dr-trigger-rb" == "$manifest_name" ]] && $bin_yq -i '.subjects[0].namespace = "{{ .Values.operator.namespace }}"' "$target_manifest"
