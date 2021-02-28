ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
SOURCE_COMMIT = $(shell git rev-parse --verify HEAD)
SOURCE_TAG = $(shell git describe --abbrev=0 --tags)
BUILD_VERSION=$(shell bv=${SOURCE_TAG} && echo $${bv:1:50})
BRANCH = $(shell git for-each-ref --format='%(objectname) %(refname:short)' refs/heads | awk "/^$$(git rev-parse HEAD)/ {print \$$2}")
COMMIT_MSG = $(shell echo "$$(git log -1 HEAD --pretty=format:%s)" | sed -e 's/'\''/"/g')
BUILD_TIME = $(shell date "+%F_%H:%M:%S")
#JWT_SECRET = $(shell < /dev/urandom tr -dc _A-Z-a-z-0-9 | head -c$${1:-24};echo;)
JWT_SECRET = TEST_JWT_SECRET
BUILD_PATH = /ntc/builds/net
BUILD_FLAGS = "-s -X main.version=${BUILD_VERSION} -X main.sourceTag=${SOURCE_TAG} -X 'main.commitMsg=${COMMIT_MSG}' -X main.gitHash=${SOURCE_COMMIT} -X 'main.branch=${BRANCH}' -X main.buildStamp=${BUILD_TIME} -X 'main.awsKey=${AWS_SECRET_ACCESS_KEY}' -X 'main.sqlPwd=${PROJ_SQL_PWD}' -X 'main.smtpPwd=${PROJ_SMTP_PWD}' -X 'main.jwtSecret=${JWT_SECRET}'"
docker_reg = ${AWS_DOCKER_REG}
registrator_token = $(shell vcli token -e dev reg)
dkname = webdav

AWS_PROFILE = gjcc
pwd = $(shell aws ecr get-login-password --region ap-northeast-1)

all: build

ecr_push: docker_build tag aws_login push

build:
	@export CGO_ENABLED=0 && go build -ldflags ${BUILD_FLAGS} -o ${BUILD_PATH}/webdav

docker_build: docker_loc_build

docker_run:
	docker run -it \
	    -p 8080:80 \
	    -e "API_PORT=80" \
        -e "VAULT_ADDR=https://ops1-1.ziongjcc.org:8200" \
        -e "VAULT_DOMAIN=ziongjcc.org" \
        -e "VAULT_FOLLOWERS=https://ops1-2.ziongjcc.org:8200,https://ops1-3.ziongjcc.org:8200" \
        -e "VAULT_REG_TOKEN=$$(vcli token -e dev reg)" \
        -e "HEBRON_PATH=/srv/media" \
        -e "HEBRON_USER=$$(jq -r '.["HEBRON_USER"]' /ntc/ops/devops/secrets/passwd/general.json)" \
        -e "HEBRON_PASSWD=$$(jq -r '.["HEBRON_PASSWD"]' /ntc/ops/devops/secrets/passwd/general.json)" \
        -e "JACOB_PATH=/srv/upload" \
        -e "JACOB_USER=$$(jq -r '.["JACOB_USER"]' /ntc/ops/devops/secrets/passwd/general.json)" \
        -e "JACOB_PASSWD=$$(jq -r '.["JACOB_PASSWD"]' /ntc/ops/devops/secrets/passwd/general.json)" \
        -e "MS_ENVIRONMENT=dev" \
        -e "WEBDAV_MODE=FULLSERVICE" \
        -e "SVC_NAME=file-us" \
        -v "/run/media/mchu/Projects/pub/media/zsf:/srv/media" \
        -v "/home/mchu/Downloads/Upload:/srv/upload" \
        webdav

docker_loc_build:
	@cp ./docker.ignore ../../.dockerignore
	@cd ../../ && \
	docker build . --file ${ROOT_DIR}/Dockerfile -t ${dkname}:${BUILD_VERSION} \
		--build-arg BRANCH=${BRANCH} \
		--build-arg SOURCE_TAG=${SOURCE_TAG} \
		--build-arg SOURCE_COMMIT=${SOURCE_COMMIT} \
		--build-arg COMMIT_MSG='${COMMIT_MSG}' \
		--build-arg BUILD_VERSION=${BUILD_VERSION} \
		--build-arg BUILD_TIME=${BUILD_TIME} \
		--build-arg AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
		--build-arg PROJ_SQL_PWD=${PROJ_SQL_PWD} \
		--build-arg PROJ_SMTP_PWD=${PROJ_SMTP_PWD} && \
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

install:
	sudo cp ${BUILD_PATH}/webdav /usr/local/bin
