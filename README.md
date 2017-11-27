# StackFoundation Sandbox

StackFoundation Sandbox is a free and open source tool that allows you to consistently reproduce complex development tasks using Docker-based workflows that are multi-step, multi-language, and multi-platform.

## Using Sandbox

If you are new to Sandbox and Docker-based workflows, visit [stack.foundation/#!/sandbox](https://stack.foundation/#!/sandbox) to learn more.

Extensive documentation on Sandbox, including quick start guides, and reference material on Sandbox and Docker-based workflows are available at [stack.foundation/#!/docs](https://stack.foundation/#!/docs). 

## Developing Sandbox

If you are interested in contributing to Sandbox, you may want to take a look at [Contributing to Sandbox](CONTRIBUTING.md) first.

### Building Sandbox

Sandbox is built on [minikube](https://github.com/kubernetes/minikube), and is written in [Go](http://golang.org/). Building Sandbox on a local OS/shell environment is done in similar way to other Go applications.

To begin, if you don't have a Go development environment, please [set one up](http://golang.org/doc/code.html). Ensure your `GOPATH` and `PATH` have been configured in accordance with the Go environment instructions.

Minikube is split into two parts: 
1) A process that runs within a linux VM called "localkube". Sandbox uses localkube as is, without any changes. That's why in order to use Sandbox, you will first need to build localkube directly from minikube.
2) Minikube itself, which is the CLI application that Sandbox is built on top of. This is the part that is contained within this repo.

You will first need to build the localkube linux binary directly from minikube.

First, clone a specific version of the minikube repo into your `GOPATH` (the repo should end up at `[GOPATH]/src/k8s.io/minikube`):

```
git clone https://github.com/kubernetes/minikube.git
cd [GOPATH]/src/k8s.io/minikube
git checkout 2deea5f75745698fd04f81da724716
```

Set the environment variable `GOOS` to `linux`, and `GOARCH` to `amd64` to build for a 64-bit Linux.
Build localkube by issuing the following command from the root of `GOPATH`:

```
go build -ldflags "-X k8s.io/minikube/pkg/version.version=v0.21.0 -X k8s.io/minikube/pkg/version.isoVersion=v0.23.1 -X k8s.io/minikube/pkg/version.isoPath=minikube/iso -s -w" k8s.io/minikube/cmd/localkube
```

Now that localkube is built, we can build Sandbox. Start by getting the Sandbox sources (again, clone into your `GOPATH`):
```
git clone https://github.com/stackfoundation/sandbox.git
cd  [GOPATH]/src/github.com/stackfoundation/sandbox
```
Get all project dependencies and "vendor" them using the `dep` tool (if you don't have the `dep`, follow the instructions [here](https://github.com/golang/dep) to install it) - run this within the locally cloned `sandbox` repo:

```
dep ensure
```

Now copy the localkube binary built from the previous step to the `core/out` folder within the `sandbox` repo. In addition, copy the `cli.zip` file from the [StackFoundation downloads](https://stack.foundation/#!/downloads) page into the `core/out` folder as well. These will be embedded into the Sandbox binary.

In order to embed the files into Sandbox, you will use a tool called `go-bindata` (find it [here](https://github.com/jteeuwen/go-bindata)) - install it by running:

```
go get -u github.com/jteeuwen/go-bindata/...
```

You will need to run the following inside of the `core` folder within the `sandbox` repo to generate a file that contains the embedded resources into the final Sandbox binary:

```
go-bindata -nomemcopy -o pkg/minikube/assets/assets.go -pkg assets ./out/localkube ./out/cli.zip deploy/addons/...
```

For macOS builds, you should also include an additional driver as an embedded artifact (from the releases page [here](https://github.com/zchee/docker-machine-driver-xhyve/releases)) and include it in the go-bindata command:

```
go-bindata -nomemcopy -o pkg/minikube/assets/assets.go -pkg assets ./out/localkube ./out/docker-machine-driver-xhyve ./out/cli.zip deploy/addons/...
```

After the project dependencies have been "vendor'ed" and the embedded resources generated, run the following in your `GOPATH` to build Sandbox _(Remember to unset the environment variables `GOOS` and `GOARCH` if you set them for building localkube)_:

```
go build github.com/stackfoundation/sandbox/core
```



