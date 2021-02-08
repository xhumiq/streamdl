module ntc.org/netutils/hosts

go 1.15

require (
	github.com/aws/aws-sdk-go v1.37.6
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.19.0
	github.com/urfave/cli/v2 v2.1.1
	gopkg.in/ini.v1 v1.62.0
	ntc.org/mclib/common v0.1.0
	ntc.org/mclib/microservice v0.1.0
	ntc.org/mclib/netutils/bitbucket v0.0.0-00010101000000-000000000000
	ntc.org/mclib/netutils/sshutils v0.1.0
)

replace (
	ntc.org/mclib/api => ../../mclib/api
	ntc.org/mclib/auth => ../../mclib/auth
	ntc.org/mclib/auth/providers => ../../mclib/auth/providers
	ntc.org/mclib/common => ../../mclib/common
	ntc.org/mclib/logger/email => ../../mclib/logger/email
	ntc.org/mclib/logger/models => ../../mclib/logger/models
	ntc.org/mclib/logger/svctail => ../../mclib/logger/svctail
	ntc.org/mclib/logger/zerolog => ../../mclib/logger/zerolog
	ntc.org/mclib/microservice => ../../mclib/microservice
	ntc.org/mclib/nechi => ../../mclib/nechi
	ntc.org/mclib/netutils/bitbucket => ../../mclib/netutils/bitbucket
	ntc.org/mclib/netutils/sshutils => ../../mclib/netutils/sshutils
	ntc.org/mclib/storage => ../../mclib/storage
	ntc.org/mclib/storage/redis => ../../mclib/storage/redis
	ntc.org/mclib/storage/sql => ../../mclib/storage/sql
)
