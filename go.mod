module sonarbridge-go

go 1.27.0

require (
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-plugin v1.8.0
	github.com/joho/godotenv v1.5.1
	github.com/nivekalara237/ci-bridge-plugin-sdk v0.0.0-20260919174840-8b9040120ff2
	github.com/spf13/cobra v1.10.2
	golang.org/x/time v0.15.0
	google.golang.org/grpc v1.84.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/fatih/color v1.19.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/oklog/run v1.2.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/net v0.59.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260918162117-cecb64721679 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

replace github.com/nivekalara237/ci-bridge-plugin-sdk => ../ci-bridge-plugin-sdk
