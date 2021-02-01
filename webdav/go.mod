module ntc.org/netutils/webdav

go 1.15

require (
	github.com/judwhite/go-svc v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.19.0
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/urfave/cli/v2 v2.1.1
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4
	golang.org/x/net v0.0.0-20190827160401-ba9fcec4b297 // indirect
	gopkg.in/yaml.v2 v2.3.0
	ntc.org/mclib/api v0.1.0
	ntc.org/mclib/common v0.1.0
	ntc.org/mclib/microservice v0.1.0
	ntc.org/mclib/nechi v0.1.0
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
	ntc.org/mclib/storage => ../../mclib/storage
	ntc.org/mclib/storage/redis => ../../mclib/storage/redis
	ntc.org/mclib/storage/sql => ../../mclib/storage/sql
)
