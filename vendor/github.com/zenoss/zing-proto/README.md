# zing-proto
gRPC and protobuf message specs for ZING, plus generated client code in go

# Contributing
1. Create a branch
2. Add/modify .proto files under the protobufs directory.  Make sure they follow the [Protobuf style guide](https://developers.google.com/protocol-buffers/docs/style)
3. Run `make clean; make`
4. Commit your changes and make a PR.

# Using
## Go projects
Add the repo to your glide.yaml with an etnry like this:
```
- package: github.com/zenoss/zing-proto/go
  version: cf7182c3e5d19ea7b6109d980f6f68b790bc2445
  subpackages:
  - metric
  - query
```
And import the desired package(s) under `github.com/zenoss/zing-proto/go`

## Java projects
Add the following dependency to your pom.xml file:
```
<dependency>
    <groupId>org.zenoss.zing.proto</groupId>
    <artifactId>zing-proto</artifactId>
    <version>1.0-SNAPSHOT</version>
</dependency>
```

And import the desired package(s) under org.zenoss.zing.proto.

# Releasing
Use git flow to release a new version to the `master` branch.

The artifact version is can be found in the [pom](java/pom.xml) file but should only be set
using the `set-devversion` (when bumping the version on develop) or `set-relversion` (when releasing to master) `make` targets.  
For example:
```
make set-devversion NEW_VERSION=1.0
```
will set the version in the pom to `1.0-SNAPSHOT` while
```
make set-relversion NEW_VERSION=1.0
```
will set the version in the pom to `1.0`.

For Zenoss employees, the details on using git-flow to release a version is documented 
on the Zenoss Engineering 
[web site](https://sites.google.com/a/zenoss.com/engineering/home/faq/developer-patterns/using-git-flow).

# References
* [Proto3 language guide](https://developers.google.com/protocol-buffers/docs/proto3)
* [Protobuf style guide](https://developers.google.com/protocol-buffers/docs/style)
* [Protobuf Go tutorial](https://developers.google.com/protocol-buffers/docs/gotutorial)
* [gRPC Go quickstart](https://grpc.io/docs/quickstart/go.html)
