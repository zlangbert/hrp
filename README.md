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

hrp is distributed as a [docker image](https://quay.io/zlangbert/hrp), making it easy to run locally, on Kubernetes, etc.

When running, you must pass a `--base-url` (repository root, `https://charts.example.com` for example) and whatever parameters are required for the chosen backend. The
web server runs on port 1323 inside the container.
  
Run with `--help` to get the full list of options:
```
docker run quay.io/zlangbert/hrp:master --help
```
  
A complete example running with the S3 backend:
```
docker run -p '1323:1323' -e 'AWS_ACCESS_KEY_ID=xxxxx' -e 'AWS_SECRET_ACCESS_KEY=xxxxx' quay.io/zlangbert/hrp:master --base-url='localhost:1323' --backend=s3 --s3-bucket=my-bucket
```

Once you have hrp running locally, you can register the repository:
```
helm repo add my-hrp http://localhost:1323
```

Push a chart:
```
curl -XPOST -F chart=@my-chart-1.2.3.tgz http://localhost:1323/chart
```

Install a chart:
```
helm install my-hrp/my-chart
```

API
=====

### `GET /index.yaml`

Returns the repository's index. This is normally used by helm itself.

```
curl http://localhost:1323/index.yaml
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

The storage backend is where the packaged charts are stored. The following is a list of the supported backends
and their configuration options.

## S3

The S3 backend stores the chart repository in an AWS S3 bucket.

#### Configuration

The only required parameter for S3 is `--s3-bucket`.

Parameters:
```
--s3-bucket=my-bucket (required)
--s3-prefix=/charts (optional)
--s3-local-sync-path=/tmp/hrp (optional)
```

A full example running the image using S3 and credentials from the local aws configuration:
```
docker run -p '1323:1323' -v $HOME/.aws:/root/.aws -e 'AWS_PROFILE=default' quay.io/zlangbert/hrp:master --base-url='localhost:1323' --backend=s3 --s3-bucket=my-bucket
```

#### Credentials

The AWS SDK is configured to use the default credentials chain. This means any standard way of consuming 
credentials will work. For example, you could mount your `.aws` folder in the container, and set `AWS_PROFILE=my-profile`,
or you could set `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` directly. If you are running on EC2 the instance profile
can also be used.
