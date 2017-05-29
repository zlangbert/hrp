package main

import (
//gopkg.in/alecthomas/kingpin.v2
)
import (
	"github.nike.com/zlangb/helm-proxy/backend"
	"github.nike.com/zlangb/helm-proxy/web"
)

func main() {

	var storage_backend backend.Backend = backend.NewS3()

	web.Start(storage_backend)
}
