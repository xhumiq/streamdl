module ntc.org/netutils/hosts

go 1.16

require (
	github.com/aws/aws-sdk-go v1.37.6
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.19.0
	github.com/urfave/cli/v2 v2.1.1
	gopkg.in/ini.v1 v1.62.0
	bitbucket.org/xhumiq/go-mclib/common v0.1.0
	bitbucket.org/xhumiq/go-mclib/microservice v0.1.0
	bitbucket.org/xhumiq/go-mclib/netutils/bitbucket v0.1.0
	bitbucket.org/xhumiq/go-mclib/netutils/linode v0.1.0
	bitbucket.org/xhumiq/go-mclib/netutils/sshutils v0.1.0
	bitbucket.org/xhumiq/go-mclib/storage v0.1.0
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
	bitbucket.org/xhumiq/go-mclib/netutils/linode => ../../mclib/netutils/linode
	bitbucket.org/xhumiq/go-mclib/netutils/sshutils => ../../mclib/netutils/sshutils
	bitbucket.org/xhumiq/go-mclib/storage => ../../mclib/storage
	bitbucket.org/xhumiq/go-mclib/storage/redis => ../../mclib/storage/redis
	bitbucket.org/xhumiq/go-mclib/storage/sql => ../../mclib/storage/sql
)
