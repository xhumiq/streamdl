module ntc.org/elzion/cicd

require (
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/go-chi/render v1.0.1
	github.com/judwhite/go-svc v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.19.0
	github.com/urfave/cli/v2 v2.1.1
	golang.org/x/text v0.3.2 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	bitbucket.org/xhumiq/go-mclib/common v0.1.0
	bitbucket.org/xhumiq/go-mclib/microservice v0.1.0
	bitbucket.org/xhumiq/go-mclib/nechi v0.1.0
)

go 1.16

replace (
	bitbucket.org/xhumiq/go-mclib/auth => ./../../mclib/auth
	bitbucket.org/xhumiq/go-mclib/common => ./../../mclib/common
	bitbucket.org/xhumiq/go-mclib/logger/email => ./../../mclib/logger/email
	bitbucket.org/xhumiq/go-mclib/logger/models => ./../../mclib/logger/models
	bitbucket.org/xhumiq/go-mclib/logger/console => ./../../mclib/logger/console
	bitbucket.org/xhumiq/go-mclib/logger/zerolog => ./../../mclib/logger/zerolog
	bitbucket.org/xhumiq/go-mclib/microservice => ./../../mclib/microservice
	bitbucket.org/xhumiq/go-mclib/nechi => ./../../mclib/nechi
	bitbucket.org/xhumiq/go-mclib/storage => ./../../mclib/storage
)
