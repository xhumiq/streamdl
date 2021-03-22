module ntc.org/netutils/webdav

go 1.16

require (
	bitbucket.org/xhumiq/go-mclib/api v0.1.0
	bitbucket.org/xhumiq/go-mclib/auth v0.1.0
	bitbucket.org/xhumiq/go-mclib/auth/cognito v0.1.0
	bitbucket.org/xhumiq/go-mclib/auth/vault v0.1.0
	bitbucket.org/xhumiq/go-mclib/common v0.1.0
	bitbucket.org/xhumiq/go-mclib/microservice v0.1.0
	bitbucket.org/xhumiq/go-mclib/nechi v0.1.0
	github.com/judwhite/go-svc v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.19.0
	github.com/urfave/cli/v2 v2.3.0
)

replace (
	bitbucket.org/xhumiq/go-mclib/api => ../../mclib/api
	bitbucket.org/xhumiq/go-mclib/auth => ../../mclib/auth
	bitbucket.org/xhumiq/go-mclib/auth/cognito => ../../mclib/auth/cognito
	bitbucket.org/xhumiq/go-mclib/auth/providers => ../../mclib/auth/providers
	bitbucket.org/xhumiq/go-mclib/auth/vault => ../../mclib/auth/vault
	bitbucket.org/xhumiq/go-mclib/common => ../../mclib/common
	bitbucket.org/xhumiq/go-mclib/logger/email => ../../mclib/logger/email
	bitbucket.org/xhumiq/go-mclib/logger/models => ../../mclib/logger/models
	bitbucket.org/xhumiq/go-mclib/logger/svctail => ../../mclib/logger/svctail
	bitbucket.org/xhumiq/go-mclib/logger/zerolog => ../../mclib/logger/zerolog
	bitbucket.org/xhumiq/go-mclib/microservice => ../../mclib/microservice
	bitbucket.org/xhumiq/go-mclib/nechi => ../../mclib/nechi
	bitbucket.org/xhumiq/go-mclib/storage => ../../mclib/storage
	bitbucket.org/xhumiq/go-mclib/storage/redis => ../../mclib/storage/redis
	bitbucket.org/xhumiq/go-mclib/storage/sql => ../../mclib/storage/sql
)
