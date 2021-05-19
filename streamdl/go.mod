module ntc.org/netutils/streamdl

require (
	bitbucket.org/xhumiq/go-mclib/auth/vault v0.1.0 // indirect
	bitbucket.org/xhumiq/go-mclib/common v0.1.0
	bitbucket.org/xhumiq/go-mclib/microservice v0.1.0
	bitbucket.org/xhumiq/go-mclib/storage v0.1.0
	github.com/judwhite/go-svc v1.1.2
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/rs/zerolog v1.19.0
	github.com/urfave/cli/v2 v2.1.1
)

go 1.16

replace (
	bitbucket.org/xhumiq/go-mclib/auth => ./../../mclib/auth
	bitbucket.org/xhumiq/go-mclib/auth/vault => ./../../mclib/auth/vault
	bitbucket.org/xhumiq/go-mclib/common => ./../../mclib/common
	bitbucket.org/xhumiq/go-mclib/logger/console => ./../../mclib/logger/console
	bitbucket.org/xhumiq/go-mclib/logger/email => ./../../mclib/logger/email
	bitbucket.org/xhumiq/go-mclib/logger/models => ./../../mclib/logger/models
	bitbucket.org/xhumiq/go-mclib/logger/zerolog => ./../../mclib/logger/zerolog
	bitbucket.org/xhumiq/go-mclib/microservice => ./../../mclib/microservice
	bitbucket.org/xhumiq/go-mclib/nechi => ./../../mclib/nechi
	bitbucket.org/xhumiq/go-mclib/storage => ./../../mclib/storage
)
