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

# for our operator cluster role, set the subject namespace
yq -i '.subjects[0].namespace = "{{ .Values.operator.namespace }}"' "$target_manifest"
