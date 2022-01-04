# warden

This is a simple command tool to help you reload the program after code is changed. You should use a yaml file to define the files to watch and scripts to run.

## Install

As for the latest version of go, you can use this command to install warden directly:

```shell
$ go install github.com/fioncat/warden@latest
```

You can alse clone this depository and install it manually:

```shell
$ git clone https://github.com/fioncat/warden.git
$ cd warden
$ make install
```

## Usage

You should create a file `.warden.yaml` under your project to configure the warden jobs. Each job describes the files that warden needs to watch and the scripts that need to be executed.

A `job` contains three stages:

- `watch`: Defines the files watched by warden.
- `build`: Defines the build scripts, warden will rerun build stage after watched files are changed.
- `exec`: Defines the main programe. This is usually a daemon process, it runs after `build` stage. Warden will kill it after watched files are changed.

For example, in a Golang Web project, we want to reload the web server after go files are changed, you can define the `.warden.yaml` like this:

```yaml
main:
  watch:
    pattern:
      - ./.../*.go  # watch all go files recursively.
  build:
    - env:
        GO111MODULE: 'on'
        CGO_ENABLED: '0'
      script: go build -o output/server main.go
  exec:
    script: ./output/server
```

Then run `warden` directly under the project:

```shell
$ warden
```

This will build and run the server, and after go files are changed, the server will be reloaded automatically.

You can pass extract args to the server by `warden` command, for example:

```shell
$ warden arg1 arg2 arg3
```

This will run `./output/server arg1 arg2 arg3` in exec stage.

Let's see a more complicated example, assuming that after the proto file is changed, we need to regenerate the rpc code file.

You can define the job like this:

```yaml
gen:
  watch:
    pattern:
      - ./api/.../*.proto  # watch proto files.
    build:
      # After proto files are changed, regenerate the rpc code.
      - script: protoc -I ./api --go-grpc-out ./rpc ./api/server.proto

main:
  watch:
    ignore:
      - api  # No need to watch 'api', it has already been watched by 'gen' job.
    pattern:
      - ./.../*.go  # watch go files
  build:
    - env:
        GO111MODULE: 'on'
        CGO_ENABLED: '0'
      script: go build -o output/server main.go
  exec:
    script: ./output/server
```

In this example, we define two jobs, the `gen` is responsible for watching the proto files and regenerating the rpc code, and the `main` is responsible for watching the go files and reloading the server.

You need to start two terminal to run these two jobs:

```shell
$ warden  # The `main` job.
```

```shell
$ warden -n gen # The `gen` job.
```

If you don't want to start two terminals, you can merge them into one job, but in order to prevent the infinite loops (will be mentioned later), you should pay special attention to configure the `ignore`:

```yaml
main:
  watch:
    ignore:
      # Ignore 'rpc' to prevent triggering another build round because of
      # the code generator changing the code files in './rpc'.
      # See the 'Infinite Loops' section for more details.
      - rpc
    pattern:
      - ./.../*.go
      - ./api/.../*.proto
  build:
    - script: protoc -I ./api --go-grpc-out ./rpc ./api/server.proto
    - env:
        GO111MODULE: 'on'
        CGO_ENABLED: '0'
      script: go build -o output/server main.go
  exec:
    script: ./output/server
```

In this way, reloading will be triggered regardless of whether the proto files or the go files are changed.

## Infinite Loop

There is a trap in the use of warden, because warden is based on file events, so if the build or exec stage modifies the watched file, a new round of build and exec will be triggered immediately, and the build and exec will continue to modify the watched file to trigger anthor round, which leads to an infinite loop.

**So, the scripts in build and exec stages MUST NOT modify the files wacthed by warden.**

This situation often occurs when some code needs to be generated during the build stage. We can reproduce this trap with the following simple example:

```yaml
main:
  watch:
    pattern:
      - ./.../*.go
  build:
    - script: rm generate.go # WARNING: This will trigger an inf loop!
    - script: touch generate.go
    - script: go build -o output/server
  exec:
    script: ./output/server
```

In the above example, when warden execute `rm generate.go; touch generate.go`, because the `generate.go` is being watched, a new round will be triggered, and this file will still be changed in the new round, so the infinite loop will happend!

To fix this, the `generate.go` file MUST BE added to ignore list:

```yaml
main:
  watch:
    ignore:
      - generate.go
    pattern:
      - ./.../*.go
  build:
    - script: rm generate.go
    - script: touch generate.go
    - script: go build -o output/server
  exec:
    script: ./output/server
```

**So if your build stage contains code generation, the generated code MUST BE added to the ignore list to prevent the infinite loop.**

You can also solve this problem by dividing the job into generating and executing job. The generating job only watch the source files of the generated code (such as proto file), and it execute the generation; the executing job watch all code files, and it execute the program.

## How it works?

Warden is mainly based on [fsnotify](https://github.com/fsnotify/fsnotify), it uses os tools to notify events after files are changed.

We only cares about `CREATE`, `WRITE`, `REMOVE`, `RENAME` events. And we don't compare files' hash code, so as long as you write the file, even if its content has not been changed, the reload will still be triggered.