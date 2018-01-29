# Hacking

Getting started with contributing or just messing around.

## Layout

The project is loosely following [golang-standards/project-layout](https://github.com/golang-standards/project-layout). The `cmd` folder has the main loops for all of the binaries produced from the project. The `internal` folder has and `app` folder, which is for application-specific code, and `pkg` for shared project code.

## Capsule8 library docs

There are no external docs yet, and no hosted godocs that I could find, so make your own:

``` bash
# they don't have a go file in the root so we have to do this in a dumb way I think
mkdir -p $GOPATH/src/github.com/capsule8
cd $GOPATH/src/github.com/capsule8
git clone https://github.com/capsule8/capsule8.git
godoc -http=127.0.0.1:6060
```

Then, hit it on localhost.

## CGO_ENABLED=0

Being able to statically compile without CGO is really nice. If something is compelling enough to consider changing this (like eBPF?) then it's worth considering. It's also worth considering making CGO features optional during the build.