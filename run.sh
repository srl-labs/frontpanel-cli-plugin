#!/usr/bin/env bash

set -o errexit
set -o pipefail

# abs path to the directory that hosts the run.sh script
BASE_DIR=$(dirname "$(readlink -f "$0")")
APPNAME=frontpanel
GOPKGNAME=${APPNAME}
BIN_DIR=${BASE_DIR}/build
BINARY=${BASE_DIR}/build/${APPNAME}
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

function golangci-lint {
	${GOLANGCI_CMD} ${GOLANGCI_FLAGS}
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
	format
	build-app
	deploy-lab
}

#################################
# Lab functions
#################################
function deploy-lab {
	containerlab deploy -c
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
