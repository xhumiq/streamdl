ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
SOURCE_COMMIT = $(shell git rev-parse --verify HEAD)
SOURCE_TAG = $(shell git describe --abbrev=0 --tags)
BUILD_VERSION=$(shell bv=${SOURCE_TAG} && echo $${bv:1:50})
BRANCH = $(shell git for-each-ref --format='%(objectname) %(refname:short)' refs/heads | awk "/^$$(git rev-parse HEAD)/ {print \$$2}")
COMMIT_MSG = $(shell echo "$$(git log -1 HEAD --pretty=format:%s)" | sed -e 's/'\''/"/g')
BUILD_TIME = $(shell date "+%F_%H:%M:%S")
BUILD_PATH = /c/ntc/builds/net
BUILD_FLAGS = "-s -X main.version=${BUILD_VERSION} -X main.sourceTag=${SOURCE_TAG} -X 'main.commitMsg=${COMMIT_MSG}' -X main.gitHash=${SOURCE_COMMIT} -X 'main.branch=${BRANCH}' -X main.buildStamp=${BUILD_TIME} -X 'main.sqlPwd=${PROJ_SQL_PWD}' -X 'main.smtpPwd=${PROJ_SMTP_PWD}'"
CGO_ENABLED = 0
GOPRIVATE=bitbucket.org/xhumiq

all: build

build:
	export GOPRIVATE=bitbucket.org/xhumiq; \
	go build -ldflags ${BUILD_FLAGS} -o ${BUILD_PATH}/streamfile.exe

install:
	sudo cp ${BUILD_PATH}/streamdl /usr/local/bin
