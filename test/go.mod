module yabl-test

go 1.17

require yabl v0.0.0

require (
	github.com/kr/pretty v0.3.0 // indirect
	github.com/rogpeppe/go-internal v1.8.1 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

require (
	github.com/gorilla/websocket v1.4.2
	gopkg.in/yaml.v2 v2.4.0
)

replace yabl => ../
