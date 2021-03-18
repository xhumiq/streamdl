FROM golang:1.16-alpine3.12 as builder
# Must run in parent folder

ARG SOURCE_COMMIT=impexport
ARG SOURCE_TAG=0.0.0
ARG COMMIT_MSG=""
ARG BUILD_VERSION=0.0.0
ARG BUILD_TIME=MISSING
ARG BRANCH=missing

ENV TIME_ZONE=Asia/Taipei
ENV GO111MODULE=on

RUN apk --no-cache add curl git tzdata bash curl-dev zlib-dev libssh2-dev libressl-dev alpine-sdk libgit2-dev && \
    cp /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && \
    echo $TIME_ZONE > /etc/timezone && \
    apk del tzdata

ENV SRC_HOME /go/src/ntc
WORKDIR $SRC_HOME

COPY ./elzion/cicd ./elzion/cicd
COPY ./mclib ./mclib

WORKDIR $SRC_HOME/elzion/cicd

RUN export BUILD_TIME=`date '+%F_%H:%M:%S'` && \
    export BUILD_FLAGS="-s -X main.version=$BUILD_VERSION -X main.sourceTag=$SOURCE_TAG -X main.gitHash=$SOURCE_COMMIT -X main.branch='$BRANCH' -X main.buildStamp=$BUILD_TIME" && \
    go build -ldflags "$BUILD_FLAGS" -o cicd

FROM node:14-alpine3.12

ENV SOURCE_COMMIT=$SOURCE_COMMIT
ENV COMMIT_MSG=$COMMIT_MSG
ENV SOURCE_TAG=$SOURCE_TAG
ENV BUILD_VERSION=$BUILD_VERSION
ENV BRANCH=$BRANCH
ENV SMTP_PASSWORD="..."
ENV BUILD_TIME=$BUILD_TIME
ENV TIME_ZONE=Asia/Taipei
ENV BUILDER_SRC_HOME /go/src/ntc
ENV API_PORT 9110

RUN apk --no-cache add tzdata ca-certificates openssh-keygen openssh-client git python2 build-base libsass libsass-dev  && \
    cp /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && \
    echo $TIME_ZONE > /etc/timezone && \
    rm -rf /var/cache/apk/* && \
    mkdir -p /root/.ssh && \
    echo $'host bb \n\
  HostName bitbucket.org \n\
  User git \n\
  IdentityFile ~/.ssh/deploy_rsa \n\
' > /root/.ssh/config && \
    echo 'bitbucket.org,18.205.93.1 ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAubiN81eDcafrgMeLzaFPsw2kNvEcqTKl/VqLat/MaB33pZy0y3rJZtnqwR2qOOvbwKZYKiEO1O6VqNEBxKvJJelCq0dTXWT5pbO2gDXC6h6QDXCaHo6pOHGPUy+YBaGQRGuSusMEASYiWunYN0vCAI8QaXnWMXNMdFP3jHAJH0eDsoiGnLPBlBp4TNm6rYI74nMzgz3B9IikW4WVK+dc8KZJZWYjAuORU3jc1c/NPskD2ASinf8v3xnfXeukU0sJ5N6m5E8VLjObPEO+mN2t/FZTMZLiFqPWc/ALSqnMnnhwrNi2rbfg/rd/IpL8Le3pSBne8+seeFVBoGqzHM9yXw==' > /root/.ssh/known_hosts

COPY --from=builder /usr/lib/lib* /usr/lib/

WORKDIR /app
COPY --from=builder $BUILDER_SRC_HOME/elzion/cicd/cicd ./
COPY --from=builder $BUILDER_SRC_HOME/elzion/cicd/bootstrap.sh ./
COPY ./elzion/cicd/configs/ /app/configs/

VOLUME ["/var/log/cicd", "/opt/elzion", "/srv/deploy/elzion", "/opt/configs/elzion", "/app/configs"]

EXPOSE $API_PORT

CMD ["/bin/sh", "-c", "/app/bootstrap.sh"]
