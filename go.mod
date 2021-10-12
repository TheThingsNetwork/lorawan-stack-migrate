module go.thethings.network/lorawan-stack-migrate

go 1.16

// Dependency of lorawan-stack.
replace gopkg.in/DATA-DOG/go-sqlmock.v1 => gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0

// Dependency of lorawan-stack.
replace gocloud.dev => gocloud.dev v0.19.0

// Dependency of lorawan-stack.
replace github.com/grpc-ecosystem/grpc-gateway => github.com/TheThingsIndustries/grpc-gateway v1.15.2-gogo

require (
	contrib.go.opencensus.io/exporter/prometheus v0.2.0 // indirect
	github.com/TheThingsNetwork/go-app-sdk v0.0.0-20191121100818-5bae20ae2b27
	github.com/TheThingsNetwork/go-utils v0.0.0-20190516083235-bdd4967fab4e
	github.com/TheThingsNetwork/ttn/core/types v0.0.0-20190516112328-fcd38e2b9dc6
	github.com/apex/log v1.1.0
	github.com/brocaar/chirpstack-api/go/v3 v3.7.5
	github.com/envoyproxy/protoc-gen-validate v0.4.1 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.7 // indirect
	github.com/jarcoal/httpmock v1.0.6 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mdempsky/unconvert v0.0.0-20200228143138-95ecdbfc0b5f
	github.com/mgechev/revive v1.0.2
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/prometheus/statsd_exporter v0.18.0 // indirect
	github.com/smartystreets/assertions v1.2.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/valyala/fasttemplate v1.2.1 // indirect
	go.thethings.network/lorawan-stack/v3 v3.15.2-0.20211011141738-8c3891a18b9f
	google.golang.org/grpc v1.37.0
	gopkg.in/ini.v1 v1.60.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)
