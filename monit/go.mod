module nex.com/telemetry/monit

go 1.12

require (
	github.com/judwhite/go-svc v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.18.0
	github.com/urfave/cli/v2 v2.1.1
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200605160147-a5ece683394c
	nex.com/nelib/api v0.1.0
	nex.com/nelib/auth/providers v0.0.0-00010101000000-000000000000 // indirect
	nex.com/nelib/common v0.1.0
	nex.com/nelib/microservice v0.1.0
	nex.com/nelib/nechi v0.1.0
)

replace (
	nex.com/nelib/api => ../../nelib/api
	nex.com/nelib/auth => ../../nelib/auth
	nex.com/nelib/auth/providers => ../../nelib/auth/providers
	nex.com/nelib/common => ../../nelib/common
	nex.com/nelib/logger/email => ../../nelib/logger/email
	nex.com/nelib/logger/models => ../../nelib/logger/models
	nex.com/nelib/logger/svctail => ../../nelib/logger/svctail
	nex.com/nelib/logger/zerolog => ../../nelib/logger/zerolog
	nex.com/nelib/microservice => ../../nelib/microservice
	nex.com/nelib/nechi => ../../nelib/nechi
	nex.com/nelib/storage => ../../nelib/storage
	nex.com/nelib/storage/redis => ../../nelib/storage/redis
	nex.com/nelib/storage/sql => ../../nelib/storage/sql
	nex.com/nelib/validation => ../../nelib/validation
)
