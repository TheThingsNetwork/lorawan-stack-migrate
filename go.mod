module go.thethings.network/lorawan-stack-migrate

go 1.14

// Dependency of lorawan-stack.
replace gopkg.in/DATA-DOG/go-sqlmock.v1 => gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0

// Dependency of lorawan-stack.
replace gocloud.dev => gocloud.dev v0.19.0

require (
	cloud.google.com/go v0.64.0 // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.2.0 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-storage-blob-go v0.10.0 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.2 // indirect
	github.com/aws/aws-sdk-go v1.34.9 // indirect
	github.com/brocaar/chirpstack-api/go/v3 v3.7.5
	github.com/envoyproxy/protoc-gen-validate v0.4.1 // indirect
	github.com/getsentry/sentry-go v0.7.0 // indirect
	github.com/go-errors/errors v1.1.1 // indirect
	github.com/go-redis/redis/v7 v7.4.0 // indirect
	github.com/gogo/protobuf v1.3.1
	github.com/google/go-replayers/grpcreplay v1.0.0 // indirect
	github.com/google/go-replayers/httpreplay v0.1.1 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20200217142428-fce0ec30dd00 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.14.7 // indirect
	github.com/jarcoal/httpmock v1.0.6 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/prometheus/common v0.13.0 // indirect
	github.com/prometheus/statsd_exporter v0.18.0 // indirect
	github.com/smartystreets/assertions v1.1.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	go.thethings.network/lorawan-stack/v3 v3.9.1
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/sys v0.0.0-20200821140526-fda516888d29 // indirect
	golang.org/x/tools v0.0.0-20200823205832-c024452afbcd // indirect
	google.golang.org/grpc v1.31.0
	gopkg.in/ini.v1 v1.60.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)
