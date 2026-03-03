#!/usr/bin/env bash

set -o errexit
set -o pipefail

# abs path to the directory that hosts the run.sh script
BASE_DIR=$(dirname "$(readlink -f "$0")")
APPNAME=frontpanel
LABDIR=${BASE_DIR}/lab
LABFILE=${APPNAME}.clab.yml

#################################
# High-Level run functions
#################################
function deploy-all {
	check-clab-version
	deploy-lab
	install-plugin
}

#################################
# Lab functions
#################################
function deploy-lab {
	mkdir -p logs/srl
	containerlab deploy -c -t ${LABDIR}
}

function destroy-lab {
	containerlab destroy -c -t ${LABDIR}/${LABFILE}
	sudo rm -rf logs/srl/*
}

function check-clab-version {
	version=$(clab version | awk '/version:/ {print $2}')
	required_version="0.68.0"
	if [[ $(echo "$version $required_version" | tr " " "\n" | sort -V | head -n 1) != "$required_version" ]]; then
		echo "Upgrade containerlab to v$required_version or newer
        Run 'sudo containerlab version upgrade' or use other installation options - https://containerlab.dev/install"
		exit 1
	fi
}

#################################
# App functions
#################################
function install-plugin {
	clab exec --label clab-node-name=frontpanel --cmd "sudo ln -s /tmp/plugin/show-${APPNAME}.py /etc/opt/srlinux/cli/plugins/show-${APPNAME}.py"
}

_run_sh_autocomplete() {
	local current_word
	COMPREPLY=()
	current_word="${COMP_WORDS[COMP_CWORD]}"

	# Get list of function names in run.sh
	local functions=$(declare -F -p | cut -d " " -f 3 | grep -v "^_" | grep -v "nvm_")

	# Generate autocompletions based on the current word
	COMPREPLY=($(compgen -W "${functions}" -- ${current_word}))
}

# Specify _run_sh_autocomplete as the source of autocompletions for run.sh
complete -F _run_sh_autocomplete ./run.sh

function help {
	printf "%s <task> [args]\n\nTasks:\n" "${0}"

	compgen -A function | grep -v "^_" | grep -v "nvm_" | cat -n

	printf "\nExtended help:\n  Each task has comments for general usage\n"
}

# This idea is heavily inspired by: https://github.com/adriancooney/Taskfile
TIMEFORMAT=$'\nTask completed in %3lR'
time "${@:-help}"
