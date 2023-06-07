module github.com/amilazy

go 1.18

require github.com/aws/amazon-ecs-agent/ecs-agent v0.0.0-20230705212230-8b82dfb78afe

replace github.com/aws/amazon-ecs-agent/ecs-agent v0.0.0-20230705212230-8b82dfb78afe => github.com/ollypom/amazon-ecs-agent/ecs-agent v0.0.0-20230707140109-71695e08457b

require (
	github.com/aws/aws-sdk-go v1.36.0 // indirect
	github.com/cihub/seelog v0.0.0-20170130134532-f561c5e57575 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/sys v0.6.0 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
