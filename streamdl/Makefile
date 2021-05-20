ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
SOURCE_COMMIT = $(shell git rev-parse --verify HEAD)
SOURCE_TAG = $(shell git describe --abbrev=0 --tags)
BUILD_VERSION=$(shell bv=${SOURCE_TAG} && echo $${bv:1:50})
BRANCH = $(shell git for-each-ref --format='%(objectname) %(refname:short)' refs/heads | awk "/^$$(git rev-parse HEAD)/ {print \$$2}")
COMMIT_MSG = $(shell echo "$$(git log -1 HEAD --pretty=format:%s)" | sed -e 's/'\''/"/g')
BUILD_TIME = $(shell date "+%F_%H:%M:%S")
BUILD_PATH = /ntc/builds/net
BUILD_FLAGS = "-s -X main.version=${BUILD_VERSION} -X main.sourceTag=${SOURCE_TAG} -X 'main.commitMsg=${COMMIT_MSG}' -X main.gitHash=${SOURCE_COMMIT} -X 'main.branch=${BRANCH}' -X main.buildStamp=${BUILD_TIME} -X 'main.sqlPwd=${PROJ_SQL_PWD}' -X 'main.smtpPwd=${PROJ_SMTP_PWD}'"
CGO_ENABLED = 0
GOPRIVATE=bitbucket.org/xhumiq
DEPLOY_VER = v0.2.1

all: build

build:
	export GOPRIVATE=bitbucket.org/xhumiq; \
	go build -ldflags ${BUILD_FLAGS} -o ${BUILD_PATH}/streamdl

install:
	cd /tmp; \
	rm -r /tmp/streamdl; \
	wget https://github.com/xhumiq/streamdl/releases/download/${DEPLOY_VER}/streamdl; \
	wget https://github.com/xhumiq/streamdl/releases/download/${DEPLOY_VER}/config.yml; \
	sudo mkdir -p /usr/lib/streamdl/${DEPLOY_VER}; \
	sudo cp /tmp/streamdl /usr/lib/streamdl/${DEPLOY_VER}/; \
	sudo rm -r /usr/bin/streamdl; \
	sudo ln -s /usr/lib/streamdl/${DEPLOY_VER}/streamdl /usr/bin/streamdl; \
	sudo mkdir -p /etc/streamdl; \
	sudo cp /tmp/config.yml /etc/streamdl/config.yml

arch_deps:
	sudo pacman -Syuuu
	sudo pacman -S ffmpeg youtube-dl

ubu_deps:
	sudo apt -y update
	sudo apt -y install ffmpeg youtube-dl
