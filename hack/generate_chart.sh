#!/bin/bash

# Copyright (c) 2023 Red Hat, Inc.

#############################################################################
###### Script for for creating a Chart from a kustomization build      ######
######                                                                 ######
###### This script uses the base manifests from the 'hack/chart_base'  ######
###### folder and a 'kustomization' build to create a chart in the     ######
###### 'hack/chart_tmp` folder (ignored by git). When done it the will ######
###### moved into the 'chart' folder for including in git.             ######
#############################################################################

# iterate over arguments and create named parameters
while [ $# -gt 0 ]; do
	if [[ $1 == *"--"* ]]; then
		param="${1/--/}"
		declare "$param"="$2"
	fi
	shift
done

# optional named parameters default values
base_folder=${temp_folder:-hack/chart_base}
target_folder=${temp_folder:-chart}
temp_folder=${temp_folder:-hack/chart_tmp}
transformers_folder=${transformers_folder:-hack/chart_transformers}
bin_yq=${bin_yq:-yq}
bin_kustomize=${bin_kustomize:-kustomize}
bin_sed=${bin_sed:-sed}
app_version=${app_version:-$(<VERSION)}
chart_version=${chart_version:-$(yq '.version' "${target_folder}"/Chart.yaml)}

# recreate the temporary folder structure (all current content will be deleted)
rm -rf "${temp_folder}"
mkdir -p "${temp_folder}"/templates
cp -rf "${base_folder}"/* "${temp_folder}"

# set temporary chart metadata values
$bin_yq -i ".version = \"${chart_version}\"" "$temp_folder"/Chart.yaml
$bin_yq -i ".appVersion = \"${app_version}\"" "$temp_folder"/Chart.yaml

# create the base templates from kustomization manifests
$bin_kustomize build config/default > "${temp_folder}"/templates/kustomized_templates.yaml
(cd "${temp_folder}"/templates && $bin_yq -s '.kind + "-" + .metadata.name' kustomized_templates.yaml)
rm -f "${temp_folder}"/templates/kustomized_templates.yaml

# utility function for injecting helm-specific labels to the manifest passed as the first argument
inject_helm_labels(){
    $bin_yq -i '.metadata.labels["helm.sh/chart"] = "{{ .Chart.Name }}-{{ .Chart.Version | replace \"+\" \"_\" }}"' "$1"
    $bin_yq -i '.metadata.labels["app.kubernetes.io/managed-by"] = "{{ .Release.Service }}"' "$1"
    $bin_yq -i '.metadata.labels["app.kubernetes.io/instance"] = "{{ .Release.Name }}"' "$1"
    $bin_yq -i '.metadata.labels["app.kubernetes.io/version"] = "{{ .Chart.AppVersion }}"' "$1"
}

# iterate over all base templates and route each to its transformer
transformers=$(find "$transformers_folder"/*.sh -maxdepth 1 -type f -printf '%f\n')
for temp_template in "$temp_folder"/templates/*.yml; do
    kind=$($bin_yq '.kind' "$temp_template")
    apiVersion=$($bin_yq '.apiVersion' "$temp_template")
    if [[ "$apiVersion" = *"/"* ]]; then
        ver_dot_group=$(tr '/' $'\n' <<< "$apiVersion" | tac | paste -s -d '.')
        transformer_name="$kind"."$ver_dot_group".sh
        # i.e. Deployment transformer_name will be Deployment.v1.apps.sh
    else
        transformer_name="$kind"."$apiVersion".sh
        # i.e. Service transformer_name will be Service.v1.sh
    fi
    for transformer in $transformers; do
        # if we have a transformer for our current kind.ver.grp / kind.ver
        if [[ "$transformer_name" == "$transformer" ]]; then
            # invoke the transformer in charge of moving content to the values file
            # and replacing content with templates on the template file
            "$transformers_folder/$transformer_name" --target_manifest "$temp_template"
            # inject helm-related labels
            inject_helm_labels "$temp_template"
            # transformers use yq to inject templates to the template file, we surround our templates with quotes to
            # suppress parsing by yq. these quotes need to be removed from the template file or they break templating
            $bin_sed -i -e "s/'//g" "$temp_template"
        fi
    done
done

# clean root chart folder and move temporary content to it
rm -rf "${target_folder:?}"/*
cp -rf "$temp_folder"/* "$target_folder"
cp -rf ./LICENSE "$target_folder"

exit 0

