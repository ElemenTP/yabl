module yabl-test

go 1.18

require yabl v0.0.0

require (
	github.com/kr/pretty v0.3.0 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

require (
	github.com/gorilla/websocket v1.5.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace yabl => ../
