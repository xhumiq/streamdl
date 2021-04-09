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
ARG BINNAME=webdav
ARG GO_SRC="/go/src"
ARG TIME_ZONE=Asia/Taipei

ENV TIME_ZONE=$TIME_ZONE

RUN apk --no-cache add curl git tzdata && \
    cp /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && \
    echo $TIME_ZONE > /etc/timezone && \
    apk del tzdata

WORKDIR $GO_SRC

COPY ./netutils/ ./netutils/
COPY ./mclib ./mclib

WORKDIR $GO_SRC/netutils/webdav

RUN export BUILD_TIME=`date '+%F_%H:%M:%S'` && \
    export JWT_SECRET=$( < /dev/urandom tr -dc _A-Z-a-z-0-9 | head -c${1:-24};echo; ) && \
    export BUILD_FLAGS="-s -X main.version=${BUILD_VERSION} -X main.sourceTag=${SOURCE_TAG} -X 'main.commitMsg=${COMMIT_MSG}' -X main.gitHash=${SOURCE_COMMIT} -X main.branch=${BRANCH} -X main.buildStamp=${BUILD_TIME}" && \
    echo $BUILD_FLAGS && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "${BUILD_FLAGS} -X 'main.awsKey=${AWS_SECRET_ACCESS_KEY}' -X 'main.sqlPwd=${PROJ_SQL_PWD}' -X 'main.smtpPwd=${PROJ_SMTP_PWD}' -X 'main.jwtSecret=${JWT_SECRET}'" -a -installsuffix cgo -o ${BINNAME}

FROM alpine:3.9

ARG SOURCE_COMMIT=chla
ARG SOURCE_TAG=0.0.0
ARG COMMIT_MSG=LOCAL_COMMIT
ARG BUILD_VERSION=0.0.0
ARG BUILD_TIME=MISSING
ARG BRANCH=MISSING
ARG BINNAME=chla
ARG GO_SRC="/go/src"
ARG TIME_ZONE=America/Los_Angeles

ENV SOURCE_COMMIT=$SOURCE_COMMIT
ENV SOURCE_TAG=$SOURCE_TAG
ENV BUILD_VERSION=$BUILD_VERSION
ENV BRANCH=$BRANCH
ENV BUILD_TIME=$BUILD_TIME
ENV TIME_ZONE=$TIME_ZONE

#postgresql-client
RUN apk --no-cache add tzdata ca-certificates bind-tools && \
    cp /usr/share/zoneinfo/$TIME_ZONE /etc/localtime && \
    echo $TIME_ZONE > /etc/timezone && \
    rm -rf /var/cache/apk/* && \
    addgroup -S webdav --gid 1499 && adduser -S webdav --uid 1499 -G webdav -h /srv/media

ENV API_PORT 4012
ENV API_TLS="false"
ENV API_CERT_PATH=""
ENV API_KEY_PATH=""
ENV LOG_FILE webdav
ENV LOG_PATH "logs"

ENV VAULT_ADDR="https://ops1-1.ziongjcc.org:8200"
ENV VAULT_DOMAIN=ziongjcc.org
ENV VAULT_FOLLOWERS="https://ops1-2.ziongjcc.org:8200,https://ops1-3.ziongjcc.org:8200"
ENV VAULT_REG_TOKEN=""
ENV HEBRON_PATH /srv/media
ENV HEBRON_USER "hebron"
ENV HEBRON_PASSWD rhema
ENV JACOB_PATH /srv/upload
ENV JACOB_USER "jacob"
ENV JACOB_PASSWD rhema
ENV MS_ENVIRONMENT=dev
ENV WEBDAV_MODE=FULLSERVICE
ENV SVC_NAME=file-us

COPY --from=builder /usr/lib/lib* /usr/lib/
WORKDIR /app

COPY --from=builder ${GO_SRC}/netutils/webdav/${BINNAME} ./
#COPY ./impexport/impexport-cli/data ./data

EXPOSE $API_PORT

CMD ["/app/webdav"]
