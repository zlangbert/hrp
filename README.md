hrp - helm repository proxy
=====

[![Docker Repository on Quay](https://quay.io/repository/zlangbert/hrp/status "Docker Repository on Quay")](https://quay.io/repository/zlangbert/hrp)
[![Build Status](https://travis-ci.org/zlangbert/hrp.svg?branch=master)](https://travis-ci.org/zlangbert/hrp)
[![Go Report Card](https://goreportcard.com/badge/github.com/zlangbert/hrp)](https://goreportcard.com/report/github.com/zlangbert/hrp)

hrp is a helm chart repository proxy with pluggable storage backends.

## Features

* acts as a helm repository and uses a storage backend for persistence
* upload charts to the repository through the HTTP API

Table of contents
=================

  * [Getting Started](#getting-started)
  * [API](#api)
  * [Backends](#backends)
    * [S3](#s3)

Getting Started
=====

```
docker run quay.io/zlangbert/hrp:master --help
```

API
=====

### `GET /index.yaml`

Returns the repository's index. This is normally used by helm itself.

```
curl http://localhost:1323/index.yaml

apiVersion: v1
entries:
  my-chart:
  - apiVersion: v1
    created: 2017-05-29T15:01:34.229784367-07:00
    description: My awesome chart in my awesome repository
    digest: e04b93e72eba81f4a2459b0d1922e5dd307e2ec7ef9a10fe9d0ecb3f675ffa96
    name: my-chart
    urls:
    - my-chart-0.1.0.tgz
    version: 0.1.0
generated: 2017-05-29T15:01:34.22940587-07:00
```

### `GET /:chart`

Download a chart, where `:chart` is of the form `my-chart-1.2.3.tgz`. This is normally used by helm itself.
 
```
curl http://localhost:1323/my-chart-1.2.3.tgz > my-chart.tgz
```

### `POST /chart`

Upload a new chart, adding it to the repository. This will replace an existing chart if that version
already exists.
 
```
curl -XPOST -F chart=@my-chart-1.2.3.tgz http://localhost:1323/chart
```


### `POST /reindex.yaml`

Forces a full reindex of the repository. If your `index.yaml` is somehow out of sync, this will regenerate it.
A reindex is automatically done on startup and when a new chart is pushed.

```
curl -XPOST http://localhost:1323/reindex
```

### `GET /health`

Returns a 200 and no content if the web server is alive.

Backends
=====

## S3
