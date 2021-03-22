ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
SOURCE_COMMIT = $(shell git rev-parse --verify HEAD)
SOURCE_TAG = $(shell git describe --abbrev=0 --tags)
BUILD_VERSION=$(shell bv=${SOURCE_TAG} && echo $${bv:1:50})
BRANCH = $(shell git for-each-ref --format='%(objectname) %(refname:short)' refs/heads | awk "/^$$(git rev-parse HEAD)/ {print \$$2}")
COMMIT_MSG = $(shell echo "$$(git log -1 HEAD --pretty=format:%s)" | sed -e 's/'\''/"/g')
BUILD_TIME = $(shell date "+%F_%H:%M:%S")
#JWT_SECRET = $(shell < /dev/urandom tr -dc _A-Z-a-z-0-9 | head -c$${1:-24};echo;)

BUILD_PATH = /c/ntc/builds/net
BUILD_FLAGS = "-s -X main.version=${BUILD_VERSION} -X main.sourceTag=${SOURCE_TAG} -X 'main.commitMsg=${COMMIT_MSG}' -X main.gitHash=${SOURCE_COMMIT} -X 'main.branch=${BRANCH}' -X main.buildStamp=${BUILD_TIME} -X 'main.awsKey=${AWS_SECRET_ACCESS_KEY}' -X 'main.sqlPwd=${PROJ_SQL_PWD}' -X 'main.smtpPwd=${PROJ_SMTP_PWD}' -X 'main.jwtSecret=${JWT_SECRET}'"

all: build

build:
	@export CGO_ENABLED=0 && go build -ldflags ${BUILD_FLAGS} -o ${BUILD_PATH}/davcli

install:
	sudo cp ${BUILD_PATH}/davcli /usr/local/bin
