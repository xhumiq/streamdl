FROM golang:1.13-alpine3.11 as builder
# Must run in parent folder

ARG SOURCE_COMMIT=idauth
ARG SOURCE_TAG=0.0.0
ARG COMMIT_MSG=""
ARG BUILD_VERSION=0.0.0
ARG BUILD_TIME=MISSING
ARG BRANCH=missing
ARG PROJ_SQL_PWD=""
ARG PROJ_SMTP_PWD=""

ENV TIME_ZONE=Asia/Taipei
ENV GO111MODULE=on

RUN apk --no-cache add curl git tzdata && \
    cp /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && \
    echo $TIME_ZONE > /etc/timezone && \
    apk del tzdata

ENV SRC_HOME /go/src/ntc
WORKDIR $SRC_HOME

COPY ./netutils/ ./netutils/
COPY ./mclib ./mclib

WORKDIR $SRC_HOME/netutils/webdav

RUN export BUILD_TIME=`date '+%F_%H:%M:%S'` && \
    export BUILD_FLAGS="-s -X main.version=$BUILD_VERSION -X main.sourceTag=$SOURCE_TAG -X 'main.commitMsg=${COMMIT_MSG}' -X main.gitHash=$SOURCE_COMMIT -X main.branch=$BRANCH -X main.buildStamp=$BUILD_TIME -X 'main.sqlPwd=${PROJ_SQL_PWD}' -X 'main.smtpPwd=${PROJ_SMTP_PWD}'" && \
    echo $BUILD_FLAGS && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "$BUILD_FLAGS" -a -installsuffix cgo -o webdav

FROM alpine:3.9

ENV SOURCE_COMMIT=$SOURCE_COMMIT
ENV SOURCE_TAG=$SOURCE_TAG
ENV BUILD_VERSION=$BUILD_VERSION
ENV BRANCH=$BRANCH
ENV BUILD_TIME=$BUILD_TIME
ENV TIME_ZONE=Asia/Taipei
ENV BUILDER_SRC_HOME /go/src/ntc

#postgresql-client
RUN apk --no-cache add tzdata ca-certificates bind-tools && \
    cp /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && \
    echo $TIME_ZONE > /etc/timezone && \
    rm -rf /var/cache/apk/* && \
    addgroup -S webdav --gid 1499 && adduser -S webdav --uid 1499 -G webdav -h /srv/media

ENV API_PORT 4012

ENV VAULT_ADDR="https://vault1.00gps.net:8200"
ENV VAULT_DOMAIN=ntc.org
ENV VAULT_FOLLOWERS=https://vault2.00gps.net,https://vault3.00gps.net
ENV VAULT_ENVIRONMENT=dev

ENV LOG_FILE webdav
ENV LOG_PATH "logs"

COPY --from=builder /usr/lib/lib* /usr/lib/
WORKDIR /app

COPY --from=builder $BUILDER_SRC_HOME/netutils/webdav/webdav ./
#COPY ./impexport/impexport-cli/data ./data

EXPOSE $API_PORT

CMD ["/app/webdav"]
