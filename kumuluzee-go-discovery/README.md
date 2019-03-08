# KumuluzEE Go Discovery

> register and discover services in kumuluzEE format.

The goal of this sample is to demonstrate how to use library for discovery in Go. The tutorial will guide you through the necessary steps. 

## Requirements

In order to run this sample you will need:
1. Go version >= 1.10.0 installed (suggested, this is version package is tested in)
    * If you have Go installed, you can check your version by typing the following in command line:
    ```
    $ go version
    ```
2. Git:
    * If you have installed Git, you can check the version by typing the following in a command line:
    ```
    $ git --version
    ```
  
## Prerequisites

We assume a working Go installation with set `$GOPATH` environment variable that points to Go's workspace.
You should know where your `$GOPATH` variable is pointing at. By default, it would be `$HOME/go` on Linux or `C:\go` on Windows.

To run this sample you will need a Consul instance. Note that such setup with only one Consul node is not viable for 
production environments, but only for developing purposes. Here is an example on how to quickly run a Consul instance with docker:
```bash
$ docker run -d --name=dev-consul --net=host consul
```

It is also recommended that you are familiar with Go programming language and it's toolchain.

## Usage

You can download these samples by running:
```bash
go get github.com/kumuluz/kumuluzee-go-samples/...
```

This will download samples and all their dependencies. Samples will be availiable in Go's workspace, under `$GOPATH/src/github.com/kumuluz/kumuluzee-go-samples`

Positioned in `.../kumuluzee-go-samples/kumuluzee-go-discovery` directory, you can run sample with command:
```bash
go run main.go
```
  
After sample is run, it can be accessed by navigating to the following URL:
* http://localhost:9000/

To shut down the example simply stop the processes in the foreground.

## Tutorial

This tutorial will guide you through the steps required to register and discover services.

We will develop a simple http server which will register service and return its url by looking it up.

### Create a Go project

Assuming working Go installation, We can create a new Go project in Go's workspace by creating a new folder in `$GOPATH/src`, for example: `$GOPATH/src/my-project`

This created project folder will serve as a root folder of our project.

### Install required dependencies

If not already, we can `go get` the *kumuluzee-go-discovery/discovery* package:
```bash
$ go get github.com/kumuluz/kumuluzee-go-discovery/discovery
```

Note that when calling the `go get` command, we should be located inside the Go's workspace.

**kumuluzee-go-discovery depends on kumuluzee-go-config, therefore when we `go get` discovery package, config package is downloaded as well.**

### Initializing discovery util

First, we are going to create file called **config.yaml**, where we will write our application's configuration:
```yaml
kumuluzee:
  # name of our service
  name: test-service
  server:
    # url where our service will live
    base-url: http://localhost:9000
    http:
      port: 9000
  env: 
    name: dev
  # specify hosts for discovery register
  discovery:
    consul:
      hosts: http://localhost:8500
  # specify hosts for remote configuration
  config:
    consul:
      hosts: http://localhost:8500
```

Now, we need to initialize our discovery util and register our service:

```go
// imports
import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "path"
    "strconv"
    "syscall"

    "github.com/kumuluz/kumuluzee-go-config/config"
    "github.com/kumuluz/kumuluzee-go-discovery/discovery"
)

var disc discovery.Util

func main() {
    // initialize discovery
    configPath := path.Join(".", "config.yaml")

    disc = discovery.New(discovery.Options{
        Extension:  "consul",
        ConfigPath: configPath,
    })

    // register service
    _, err := disc.RegisterService(discovery.RegisterOptions{})
    if err != nil {
        panic(err)
    }

    // perform service deregistration on received interrupt or terminate signals
    deregisterOnSignal()

    // here we also add http server, see below
}
```

Note that we also have to make sure to **deregister** service once it stops working. With Go, we can use `os/signal` standard package to handle received signals (i.e. interrupt and terminate signals):

```go
func deregisterOnSignal() {
    // catch interrupt or terminate signals and send them to sigs channel
    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

    // function waits for received signal - and then performs service deregistration
    go func() {
        <-sigs
        if err := disc.DeregisterService(); err != nil {
            panic(err)
        }
        // besides deregistration, you could also do any other clean-up here.
        // Make sure to call os.Exit() with status number at the end.
        os.Exit(1)
    }()
}
```

Now we will create simple http server, which will lookup url of our service and send it to client:
```go
http.HandleFunc("/lookup", func(w http.ResponseWriter, r *http.Request) {
    // define parameters of the service we are looking for
    // and call DiscoverService
    service, err := disc.DiscoverService(discovery.DiscoverOptions{
        Value:       "test-service",
        Version:     "1.0.0",
        Environment: "dev",
        AccessType:  "direct",
    })
    if err != nil {
        w.WriteHeader(500)
        fmt.Fprint(w, err.Error())
    } else {
        // prepare a struct for marshalling into json
        data := struct {
            Service string `json:"service"`
        }{
            serviceURL,
        }

        // generate json from data
        genjson, err := json.Marshal(data)
        if err != nil {
            w.WriteHeader(500)
        } else {
            // write generated json to ResponseWriter
            fmt.Fprint(w, string(genjson))
        }
    }
})

log.Fatal(http.ListenAndServe(":9000", nil))
```

Upon visiting http://localhost:9000/lookup , the response should be:
```json
{"service":"127.0.0.1:9000"}
```