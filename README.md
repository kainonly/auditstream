# Weplanx PPCollector (Periodic Polling Collector)

[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/weplanx/ppcollector/release.yml?label=release&style=flat-square)](https://github.com/weplanx/ppcollector/actions/workflows/release.yml)
[![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/weplanx/ppcollector/testing.yml?label=testing&style=flat-square)](https://github.com/weplanx/ppcollector/actions/workflows/testing.yml)
[![Release](https://img.shields.io/github/v/release/weplanx/ppcollector.svg?style=flat-square&include_prereleases)](https://github.com/weplanx/ppcollector/releases)
[![Coveralls github](https://img.shields.io/coveralls/github/weplanx/ppcollector.svg?style=flat-square)](https://coveralls.io/github/weplanx/ppcollector)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/weplanx/ppcollector?style=flat-square)](https://github.com/weplanx/ppcollector)
[![Go Report Card](https://goreportcard.com/badge/github.com/weplanx/ppcollector?style=flat-square)](https://goreportcard.com/report/github.com/weplanx/ppcollector)
[![GitHub license](https://img.shields.io/github/license/weplanx/ppcollector?style=flat-square)](https://raw.githubusercontent.com/weplanx/ppcollector/main/LICENSE)

A streamlined, professional queue-based data collector tailored for ClickHouse time-series storage, designed to leverage periodic polling for efficient batch writes from the queue.

## Pre-requisite

- A NATS JetStream cluster is required.
- A ClickHouse database is required, with the version as high as possible.
- The transfer and collector must use the same NATS cluster, and the same application namespace.

## Deploy

A collector service that subscribes to stream queues and then writes to data.

The main container image is:

- ghcr.io/weplanx/ppcollector:latest

The case will use Kubernetes deployment orchestration, replicate deployment (modify as needed).

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ppcollector
spec:
  selector:
    matchLabels:
      app: ppcollector
  template:
    metadata:
      labels:
        app: ppcollector
    spec:
      containers:
        - image: ghcr.io/weplanx/ppcollector:latest
          imagePullPolicy: Always
          name: ppcollector
```

## License

[BSD-3-Clause License](https://github.com/weplanx/ppcollector/blob/main/LICENSE)
