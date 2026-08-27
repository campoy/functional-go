module github.com/campoy/functional-go/fpgen

go 1.27

require (
	github.com/campoy/functional-go v0.0.0
	github.com/stretchr/testify v1.12.1
)

require go.yaml.in/yaml/v3 v3.0.5 // indirect

replace github.com/campoy/functional-go => ../
