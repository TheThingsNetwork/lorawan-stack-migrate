module go.thethings.network/lorawan-stack-migrate

go 1.18

// Dependency of lorawan-stack.
replace gopkg.in/DATA-DOG/go-sqlmock.v1 => gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0

// Dependency of lorawan-stack.
replace gocloud.dev => gocloud.dev v0.19.0

// Dependency of lorawan-stack.
replace github.com/grpc-ecosystem/grpc-gateway => github.com/TheThingsIndustries/grpc-gateway v1.15.2-gogo

// Dependency of lorawan-stack.
replace github.com/grpc-ecosystem/grpc-gateway/v2 => github.com/grpc-ecosystem/grpc-gateway/v2 v2.10.3

// Dependency of lorawan-stack.
replace google.golang.org/grpc => google.golang.org/grpc v1.33.1

require (
	github.com/TheThingsNetwork/go-app-sdk v0.0.0-20191121100818-5bae20ae2b27
	github.com/TheThingsNetwork/go-utils v0.0.0-20200807125606-b3493662e4bf
	github.com/TheThingsNetwork/ttn/core/types v0.0.0-20190516112328-fcd38e2b9dc6
	github.com/apex/log v1.9.0
	github.com/brocaar/chirpstack-api/go/v3 v3.12.5
	github.com/gogo/protobuf v1.3.2
	github.com/mdempsky/unconvert v0.0.0-20200228143138-95ecdbfc0b5f
	github.com/mgechev/revive v1.2.3
	github.com/smartystreets/assertions v1.13.0
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	go.thethings.network/lorawan-stack/v3 v3.21.2
	go.uber.org/zap v1.23.0
	google.golang.org/grpc v1.49.0
)

require (
	contrib.go.opencensus.io/exporter/prometheus v0.4.0 // indirect
	github.com/BurntSushi/toml v1.2.0 // indirect
	github.com/TheThingsIndustries/protoc-gen-go-flags v1.0.0 // indirect
	github.com/TheThingsIndustries/protoc-gen-go-json v1.4.0 // indirect
	github.com/TheThingsNetwork/api v0.0.0-20190516111443-a3523f89e84f // indirect
	github.com/TheThingsNetwork/go-account-lib v0.0.0-20190516094738-77d15a3f8875 // indirect
	github.com/TheThingsNetwork/ttn/api v0.0.0-20190516081709-034d40b328bd // indirect
	github.com/TheThingsNetwork/ttn/mqtt v0.0.0-20190516112328-fcd38e2b9dc6 // indirect
	github.com/TheThingsNetwork/ttn/utils/errors v0.0.0-20190516081709-034d40b328bd // indirect
	github.com/TheThingsNetwork/ttn/utils/random v0.0.0-20190516092602-86414c703ee1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/brocaar/lorawan v0.0.0-20170626123636-a64aca28516d // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/chavacava/garif v0.0.0-20220630083739-93517212f375 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/eclipse/paho.mqtt.golang v1.3.5 // indirect
	github.com/envoyproxy/protoc-gen-validate v0.6.3 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/go-kit/log v0.2.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/mock v1.4.4 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/gotnospirit/makeplural v0.0.0-20180622080156-a5f48d94d976 // indirect
	github.com/gotnospirit/messageformat v0.0.0-20190719172517-c1d0bdacdea2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.0.0-00010101000000-000000000000 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jacobsa/crypto v0.0.0-20190317225127-9f44e2d11115 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mgechev/dots v0.0.0-20210922191527-e955255bf517 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/mwitkow/go-grpc-middleware v1.0.0 // indirect
	github.com/oklog/ulid/v2 v2.0.2 // indirect
	github.com/olekukonko/tablewriter v0.0.5 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/prometheus/statsd_exporter v0.22.4 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	gocloud.dev v0.20.0 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/exp v0.0.0-20230108222341-4b8118a2686a // indirect
	golang.org/x/mod v0.6.0 // indirect
	golang.org/x/net v0.1.0 // indirect
	golang.org/x/oauth2 v0.0.0-20220411215720-9780585627b5 // indirect
	golang.org/x/sys v0.1.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	golang.org/x/tools v0.2.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220519153652-3a47de7e79bd // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/redis.v5 v5.2.9 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
