# service gateway

[![reportcard](https://goreportcard.com/badge/github.com/gomatic/gateway)](https://goreportcard.com/report/github.com/gomatic/gateway)
[![build](https://travis-ci.org/gomatic/gateway.svg?branch=master)](https://travis-ci.org/gomatic/gateway)
[![godoc](https://godoc.org/github.com/gomatic/gateway?status.svg)](https://godoc.org/github.com/gomatic/gateway)
[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](http://www.gnu.org/licenses/gpl-3.0)

## Install

    go install github.com/gomatic/gateway

## Test

**Install the example service**

    go install github.com/gomatic/service-example

**Run the gateway**

    EXAMPLE_SERVICE_PORT=5000 gateway --debug >gateway.log 2>&1 & gateway_pid=$!

**Test the gateway**

    curl -s localhost:3000/health         # The health check

debug routes

    curl -s localhost:2999/header         # This debug route generates a JWT token
    curl -s localhost:2999/debug/vars     # This debug route provides runtime information

**Run the example service**

    API_PORT=5000 service-example --debug >service-example.log 2>&1 & service_pid=$!

**Test the example service**

    curl -s localhost:5000/health         # The health check
    curl -s localhost:5000/api/model.json # The OpenAPI documentation

debug routes

    curl -s localhost:4999/debug/vars     # This debug route provides runtime information

**Call the example service through the gateway**

    curl -is -H "$(curl -s localhost:2999/header)" http://localhost:3000/v1/example/health
    curl -is -H "$(curl -s localhost:2999/header)" http://localhost:3000/v1/example/api/model.json

### RPC

**Install the example RPC client**

    go install github.com/gomatic/service-example/cmd/service-example-client

**Call the example service through the RPC using the client helper**

    API_PORT=5000 service-example-client this is a great example message

### Cleanup

    kill ${gateway_pid} ${service_pid}
    rm gateway.log service-example.log
