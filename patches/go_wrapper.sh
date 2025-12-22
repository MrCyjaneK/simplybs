#!/bin/bash
export GOCACHE=$HOME/.cache/go@GO_VERSION@-build
export GOMODCACHE=$HOME/go@GO_VERSION@/pkg/mod
export GOPATH=$HOME/go@GO_VERSION@
export GOPROXY='https://proxy.golang.org,direct'
export GOROOT=$PREFIX/native/go@GO_VERSION@
export GOSUMDB='sum.golang.org'
export GOTELEMETRY='local'
export GOTELEMETRYDIR='/Users/user/Library/Application Support/go/telemetry'
export GOTOOLDIR=$GOROOT/pkg/tool/${BUILDER_GOOS}_${BUILDER_GOARCH}

exec $PREFIX/native/go@GO_VERSION@/bin/go $@
