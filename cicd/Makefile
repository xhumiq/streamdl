ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
SOURCE_COMMIT = $(shell git rev-parse --verify HEAD)
SOURCE_TAG = $(shell git describe --abbrev=0 --tags)
BUILD_VERSION=$(shell bv=${SOURCE_TAG} && echo $${bv:1:50})
BRANCH = $(shell git for-each-ref --format='%(objectname) %(refname:short)' refs/heads | awk "/^$$(git rev-parse HEAD)/ {print \$$2}")
COMMIT_MSG = $(shell echo "$$(git log -1 HEAD --pretty=format:%s)" | sed -e 's/'\''/"/g')
BUILD_TIME = $(shell date "+%F_%H:%M:%S")
BUILD_PATH = /ntc/builds/elzion
BUILD_FLAGS = "-s -X main.version=${BUILD_VERSION} -X main.sourceTag=${SOURCE_TAG} -X 'main.commitMsg=${COMMIT_MSG}' -X main.gitHash=${SOURCE_COMMIT} -X 'main.branch=${BRANCH}' -X main.buildStamp=${BUILD_TIME} -X 'main.sqlPwd=${PROJ_SQL_PWD}' -X 'main.smtpPwd=${PROJ_SMTP_PWD}'"
docker_reg = ${AWS_DOCKER_REG}
dkname = cicd

AWS_PROFILE = gjcc
pwd = $(shell aws ecr get-login-password --region ap-northeast-1)

all: build

ecr_push: aws_login docker_build tag push

build:
    CGO_ENABLED = 0 && go build -ldflags ${BUILD_FLAGS} -o ${BUILD_PATH}/hosts

docker_build:
	cp ./docker.ignore ../../.dockerignore
	cd ../../ && \
	docker build . --file ${ROOT_DIR}/Dockerfile -t ${dkname}:${BUILD_VERSION} \
		--build-arg BRANCH=${BRANCH} \
		--build-arg SOURCE_TAG=${SOURCE_TAG} \
		--build-arg SOURCE_COMMIT=${SOURCE_COMMIT} \
		--build-arg COMMIT_MSG=${COMMIT_MSG} \
		--build-arg BUILD_VERSION=${BUILD_VERSION} \
		--build-arg BUILD_TIME=${BUILD_TIME} && \
		docker tag ${dkname}:${BUILD_VERSION} ${dkname}:latest && \
		docker tag ${dkname}:$(BUILD_VERSION) ${dkname} && \
		echo Set ${dkname}:$(BUILD_VERSION) to latest && \
		echo "SOURCE_COMMIT=${SOURCE_COMMIT}" && \
		echo "SOURCE_TAG=   ${SOURCE_TAG}" && \
		echo "COMMIT_MSG=   ${COMMIT_MSG}" && \
		echo "BUILD_VERSION=${BUILD_VERSION}" && \
		echo "BUILD_TIME=   ${BUILD_TIME}" && \
		echo "BRANCH=       ${BRANCH}"

aws_login:
	@docker login --username AWS --password ${pwd} ${docker_reg}

tag:
	docker tag ${dkname} ${docker_reg}/${dkname}:${BUILD_VERSION} && \
	docker tag ${dkname} ${docker_reg}/${dkname}:latest

push:
	docker push ${docker_reg}/${dkname}:${BUILD_VERSION}
	docker push ${docker_reg}/${dkname}:latest
