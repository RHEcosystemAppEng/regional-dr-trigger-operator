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

# set namespace selector for prometheus service monitor
$bin_yq -i '.spec.namespaceSelector.matchNames[0] = "{{ .Values.operator.namespace }}"' "$target_manifest"
