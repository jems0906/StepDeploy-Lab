module github.com/example/stepdeploy-lab/services/ca-service

go 1.22

require (
	github.com/example/stepdeploy-lab v0.0.0
	github.com/gorilla/mux v1.8.1
)

replace github.com/example/stepdeploy-lab => ../..
