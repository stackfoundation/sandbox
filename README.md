# StackFoundation Sandbox

StackFoundation Sandbox is a free and open source tool that allows you to consistently reproduce complex development tasks using Docker-based workflows that are multi-step, multi-language, and multi-platform.

## Using Sandbox

If you are new to Sandbox and Docker-based workflows, visit [stack.foundation/#!/sandbox](https://stack.foundation/#!/sandbox) to learn more.

Extensive documentation on Sandbox, including quick start guides, and reference material on Sandbox and Docker-based workflows are available at [stack.foundation/#!/docs](https://stack.foundation/#!/docs). 

## Developing Sandbox

### Building Sandbox

Sandbox is built on [minikube](https://github.com/kubernetes/minikube), and is written in [Go](http://golang.org/). Building Sandbox on a local OS/shell environment is done in similar way to other Go applications.

To begin, if you don't have a Go development environment, please [set one up](http://golang.org/doc/code.html). Ensure your GOPATH and PATH have been configured in accordance with the Go environment instructions.

Minikube is split into two parts: 1) a process that runs within a linux VM called "localkube", and minikube itself, which is the CLI application that Sandbox is built on top of. You will first need to build the localkube linux binary directly from minikube.

Get a specific version of minikube sources:

```
git clone https://github.com/kubernetes/minikube.git
cd minikube
git checkout 2deea5f75745698fd04f81da724716
```

Set the environment variable `GOOS` to `linux`, and `GOARCH` to `amd64` to build for a 64-bit linux.
Build localkube by issuing the following command at your GOPATH root:

```
go build k8s.io/minikube/cmd/localkube
```

Now that localkube is built, we can build Sandbox. Start by getting the Sandbox sources:
```
git clone https://github.com/stackfoundation/sandbox.git
cd sandbox
```
Get all project dependencies and "vendor" them using the `dep` tool (if you don't have the `dep`, follow the instructions [here](https://github.com/golang/dep) to install it) - this will only have to be done when dependencies change:

```
dep ensure
```

Now copy the localkube binary to the `core/out` folder within the Sandbox repository. In addition, copy the `cli.zip` file from the [StackFoundation downloads](https://stack.foundation/#!/downloads) page into the `core/out` folder as well. These will be embedded into the Sandbox binary.

You will need to run the following to generate a file that contains the embedded resources into the final Sandbox binary:

```
go get -u github.com/jteeuwen/go-bindata/...
go-bindata -nomemcopy -o pkg/minikube/assets/assets.go -pkg assets ./out/localkube ./out/cli.zip deploy/addons/...
```

For macOS builds, you should also include an additional driver as an embedded artifact (from the releases page [here](https://github.com/zchee/docker-machine-driver-xhyve/releases)):

```
go-bindata -nomemcopy -o pkg/minikube/assets/assets.go -pkg assets ./out/localkube ./out/docker-machine-driver-xhyve ./out/cli.zip deploy/addons/...
```

After the project dependencies have been "vendor'ed" and the embedded resources generated, run the following to build Sandbox:

```
go build github.com/stackfoundation/sandbox/core
```



