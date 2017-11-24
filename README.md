# StackFoundation Sandbox

StackFoundation Sandbox is a free and open source tool that allows you to consistently reproduce complex development tasks using Docker-based workflows that are multi-step, multi-language, and multi-platform.

## Using Sandbox

If you are new to Sandbox and Docker-based workflows, visit [stack.foundation/#!/sandbox](https://stack.foundation/#!/sandbox) to learn more.

Extensive documentation on Sandbox, including quick start guides, and reference material on Sandbox and Docker-based workflows are available at [stack.foundation/#!/docs](https://stack.foundation/#!/docs). 

## Developing Sandbox

### Building Sandbox

Sandbox is built on [minikube](https://github.com/kubernetes/minikube), and is written in [Go](http://golang.org/). Building Sandbox on a local OS/shell environment is done in similar way to other Go applications.

To begin, if you don't have a Go development environment, please [set one up](http://golang.org/doc/code.html). Ensure your GOPATH and PATH have been configured in accordance with the Go environment instructions.

Get the Sandbox sources:
```
git clone https://github.com/stackfoundation/sandbox.git
cd sandbox
```
Get all project dependencies and "vendor" them using the `dep` tool (if you don't have the `dep`, follow the instructions [here](https://github.com/golang/dep) to install it) - this will only have to be done once:

```
dep ensure
```

You will need to run the following to generate a file that contains resources that are embedded into the final Sandbox binary:

```
go get -u github.com/jteeuwen/go-bindata/...
go-bindata -nomemcopy -o pkg/minikube/assets/assets.go -pkg assets ./out/localkube ./out/cli.zip deploy/addons/...
```

For macOS builds, you should also include an additional driver as an embedded artifact:

```
go-bindata -nomemcopy -o pkg/minikube/assets/assets.go -pkg assets ./out/localkube ./out/docker-machine-driver-xhyve ./out/cli.zip deploy/addons/...
```

After the project dependencies have been "vendor'ed" and the embedded resources generated, run the following to build Sandbox:

```
go build github.com/stackfoundation/sandbox/core
```



