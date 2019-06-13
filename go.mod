module github.com/criteo-forks/azure_metrics_exporter

go 1.12

replace github.com/RobustPerception/azure_metrics_exporter => ./

require (
	contrib.go.opencensus.io/exporter/ocagent v0.2.0
	github.com/Azure/azure-sdk-for-go v21.3.0+incompatible
	github.com/Azure/go-autorest v11.2.7+incompatible
	github.com/RobustPerception/azure_metrics_exporter v0.0.0-00010101000000-000000000000
	github.com/alecthomas/template v0.0.0-20160405071501-a0175ee3bccc
	github.com/alecthomas/units v0.0.0-20151022065526-2efee857e7cf
	github.com/beorn7/perks v1.0.0
	github.com/census-instrumentation/opencensus-proto v0.1.0
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dimchansky/utfbom v1.0.0
	github.com/golang/protobuf v1.3.1
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/mitchellh/go-homedir v1.0.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v0.9.4
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90
	github.com/prometheus/common v0.4.1
	github.com/prometheus/procfs v0.0.2
	go.opencensus.io v0.22.0
	golang.org/x/crypto v0.0.0-20190611184440-5c40567a22f8
	golang.org/x/net v0.0.0-20190611141213-3f473d35a33a
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	golang.org/x/sys v0.0.0-20190613124609-5ed2794edfdc
	golang.org/x/text v0.3.2
	google.golang.org/api v0.6.0
	google.golang.org/genproto v0.0.0-20190611190212-a7e196e89fd3
	google.golang.org/grpc v1.21.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v2 v2.2.2
)
