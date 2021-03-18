module ntc.org/netutils/textutils

go 1.16

require (
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351
	github.com/rs/zerolog v1.19.0
	github.com/urfave/cli/v2 v2.1.1
	golang.org/x/text v0.3.3
	bitbucket.org/xhumiq/go-mclib/auth v0.1.0
	bitbucket.org/xhumiq/go-mclib/common v0.1.0
	bitbucket.org/xhumiq/go-mclib/microservice v0.1.0
	bitbucket.org/xhumiq/go-mclib/netutils/bitbucket v0.1.0
	bitbucket.org/xhumiq/go-mclib/netutils/sshutils v0.1.0
)

replace (
	bitbucket.org/xhumiq/go-mclib/api => ../../mclib/api
	bitbucket.org/xhumiq/go-mclib/auth => ../../mclib/auth
	bitbucket.org/xhumiq/go-mclib/auth/providers => ../../mclib/auth/providers
	bitbucket.org/xhumiq/go-mclib/common => ../../mclib/common
	bitbucket.org/xhumiq/go-mclib/logger/email => ../../mclib/logger/email
	bitbucket.org/xhumiq/go-mclib/logger/models => ../../mclib/logger/models
	bitbucket.org/xhumiq/go-mclib/logger/svctail => ../../mclib/logger/svctail
	bitbucket.org/xhumiq/go-mclib/logger/zerolog => ../../mclib/logger/zerolog
	bitbucket.org/xhumiq/go-mclib/microservice => ../../mclib/microservice
	bitbucket.org/xhumiq/go-mclib/nechi => ../../mclib/nechi
	bitbucket.org/xhumiq/go-mclib/netutils/bitbucket => ../../mclib/netutils/bitbucket
	bitbucket.org/xhumiq/go-mclib/netutils/sshutils => ../../mclib/netutils/sshutils
	bitbucket.org/xhumiq/go-mclib/storage => ../../mclib/storage
	bitbucket.org/xhumiq/go-mclib/storage/redis => ../../mclib/storage/redis
	bitbucket.org/xhumiq/go-mclib/storage/sql => ../../mclib/storage/sql
)
