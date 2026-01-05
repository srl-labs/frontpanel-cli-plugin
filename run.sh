#!/usr/bin/env bash

set -o errexit
set -o pipefail

# abs path to the directory that hosts the run.sh script
BASE_DIR=$(dirname "$(readlink -f "$0")")
APPNAME=frontpanel
GOPKGNAME=${APPNAME}
BIN_DIR=${BASE_DIR}/build
BINARY=${BASE_DIR}/build/${APPNAME}
LABDIR=${BASE_DIR}/lab
LABFILE=${APPNAME}.clab.yml

GOLANGCI_CMD="sudo docker run -t --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.60.3 golangci-lint"
GOLANGCI_FLAGS="run -v ./..."

GOIMPORTS_CMD="sudo docker run --rm -it -v $(pwd):/work -w /work ghcr.io/hellt/goimports:v0.25.0"
GOIMPORTS_FLAGS="-w ."

COMMON_LDFLAGS="-X main.version=dev -X main.commit=$(git rev-parse --short HEAD)"

GOMPLATE_IMAGE="ghcr.io/hairyhenderson/gomplate:v4.3-alpine"
YANGLINT_IMAGE="ghcr.io/hellt/yanglint:3.7.8"

if [ -z "$NDK_DEBUG" ]; then
	# when not in debug mode use linker flags -s -w to strip the binary
	LDFLAGS="-s -w $COMMON_LDFLAGS\""
else
	# when NDK_DEBUG is set
	LDFLAGS="$COMMON_LDFLAGS"
	GCFLAGS="all=-N -l"

	# links the dlv binary to the debug directory as a hardlink
	# making it available to the frontpanel container when running in debug mode.
	ln -f $(which dlv) ${BASE_DIR}/debug/
fi

#################################
# Build and lint functions
#################################
function lint-yang {
	echo "Linting YANG files"
	#docker run --rm -v ${BASE_DIR}:/work ${YANGLINT_IMAGE} yang/*.yang
}

function lint-yaml {
	echo "Linting YAML files"
	docker run --rm -v ${BASE_DIR}/${APPNAME}.yml:/data/${APPNAME}.yml cytopia/yamllint -d relaxed .
}

function golangci-lint {
	${GOLANGCI_CMD} ${GOLANGCI_FLAGS}
}

function lint {
	lint-yang
	lint-yaml
}

GOFUMPT_CMD="docker run --rm -it -e GOFUMPT_SPLIT_LONG_LINES=on -v ${BASE_DIR}:/work ghcr.io/hellt/gofumpt:v0.7.0"
GOFUMPT_FLAGS="-l -w ."

GODOT_CMD="docker run --rm -it -v ${BASE_DIR}:/work ghcr.io/hellt/godot:1.4.11"
GODOT_FLAGS="-w ."

function gofumpt {
	${GOFUMPT_CMD} ${GOFUMPT_FLAGS}
}

function godot {
	${GODOT_CMD} ${GODOT_FLAGS}
}

function goimports {
	${GOIMPORTS_CMD} ${GOIMPORTS_FLAGS}
}

function format {
	goimports
	gofumpt
	godot
	# format the run.sh file
	sudo docker run --rm -u "$(id -u):$(id -g)" -v "$(pwd):/mnt" -w /mnt mvdan/shfmt:latest -l -w run.sh >/dev/null
}

function build-app {
	lint
	format
	echo "Building application"
	mkdir -p ${BIN_DIR}
	go mod tidy

	if [[ -n "${NDK_DEBUG}" ]]; then
		go build -race -o ${BINARY} -ldflags="${LDFLAGS}" -gcflags="${GCFLAGS}" .
	else
		go build -o ${BINARY} -ldflags="${LDFLAGS}" -gcflags="${GCFLAGS}" .
	fi
}

#################################
# High-Level run functions
#################################
function deploy-all {
	check-clab-version
	template-files
	build-app
	deploy-lab
	install-app
}

# This function is used to re-deploy the app
# without re-deploying the lab
# The workflow is:
# 1. first deploy the lab with `./run.sh deploy-all`
# 2. make changes to the app code
# 3. run `./run.sh build-restart-app`
# which will rebuild the app and restart it without
# requiring to re-deploy the lab
function build-restart-app {
	build-app
	reload-app_mgr
	sleep 3 # wait 3s for app_mgr to reload
	restart-app
}

function template-files {
	template-lab
	template-app
}

#################################
# Lab functions
#################################
function deploy-lab {
	mkdir -p logs/srl
	mkdir -p logs/frontpanel
	containerlab deploy -c -t ${LABDIR}
}

function destroy-lab {
	containerlab destroy -c -t ${LABDIR}/${LABFILE}
	sudo rm -rf logs/srl/* logs/frontpanel/*
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

# template lab file. When NDK_DEBUG env var is set to any value
# the debug section is added to the lab file
function template-lab {
	echo "Templating lab file"
	sudo docker run --rm -e NDK_DEBUG=${NDK_DEBUG} -v ${BASE_DIR}/lab/:/tmp/ \
		${GOMPLATE_IMAGE} \
		--file /tmp/${LABFILE}.go.tpl -o /tmp/${LABFILE}
}

#################################
# App functions
#################################

# template app file. When NDK_DEBUG env var is set to any value
# the debug section is added to the app configuration file.
function template-app {
	echo "Templating app file"
	sudo docker run --rm -e NDK_DEBUG=${NDK_DEBUG} -e NOWAIT=${NOWAIT} \
		-v ${BASE_DIR}:/tmp \
		${GOMPLATE_IMAGE} \
		--file /tmp/${APPNAME}.yml.go.tpl -o /tmp/${APPNAME}.yml
}

# install-app creates app symlinks and reloads app_mgr
# which effectively installs and starts the app as app_mgr
# becomes aware of it
# this technique is used so that we can re-build the app later
# and have the new binary picked up by app_mgr without redeploying the lab
function install-app {
	create-app-symlink
	reload-app_mgr
}

function show-app-status {
	clab exec --label containerlab=frontpanel --cmd "sr_cli show system application ${APPNAME}"
}

function restart-app {
	clab exec --label containerlab=frontpanel --cmd "sr_cli tools system app-management application ${APPNAME} restart"
}

function reload-app {
	clab exec --label containerlab=frontpanel --cmd "sr_cli tools system app-management application ${APPNAME} reload"
}

function stop-app {
	clab exec --label containerlab=frontpanel --cmd "sr_cli tools system app-management application ${APPNAME} stop"
}

function start-app {
	clab exec --label containerlab=frontpanel --cmd "sr_cli tools system app-management application ${APPNAME} start"
}

function redeploy-app {
	build-app
	reload-app
}

function create-app-symlink {
	clab exec --label containerlab=frontpanel --cmd "sudo ln -s /tmp/build/${APPNAME} /usr/local/bin/${APPNAME}"
	clab exec --label containerlab=frontpanel --cmd "sudo ln -s /tmp/plugin/show-${APPNAME}.py /etc/opt/srlinux/cli/plugins/show-${APPNAME}.py"
	clab exec --label containerlab=frontpanel --cmd "sudo ln -s /tmp/${APPNAME}.yml /etc/opt/srlinux/appmgr/${APPNAME}.yml"
}

function reload-app_mgr {
	clab exec --label containerlab=frontpanel --cmd "sr_cli tools system app-management application app_mgr reload"
}

#################################
# Packaging functions
#################################
function compress-bin {
	rm -f build/compressed
	chmod 777 build/${APPNAME}
	docker run --rm -v $(pwd):/work ghcr.io/hellt/upx:4.0.2-r0 --best --lzma -o build/compressed build/${APPNAME}
	mv build/compressed build/${APPNAME}
}

# package packages the binary into a deb package by default
# if `rpm` is passed as an argument, it will create an rpm package
function package {
	build-app
	compress-bin
	local packager=${1:-deb}
	docker run --rm -v $(pwd):/tmp -w /tmp ghcr.io/goreleaser/nfpm:v2.40.0 package \
		--config /tmp/nfpm.yml \
		--target /tmp/build \
		--packager ${packager}
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
