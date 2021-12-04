# kuttilib

An API to manage Kubernetes clusters and nodes.

[![Go Report Card](https://goreportcard.com/badge/github.com/kuttiproject/kuttilib)](https://goreportcard.com/report/github.com/kuttiproject/kuttilib)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/kuttiproject/kuttilib)](https://pkg.go.dev/github.com/kuttiproject/kuttilib)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/kuttiproject/kuttilib?include_prereleases)

The kutti project aims to let users create small, non production Kubernetes clusters for learning and development. The kuttilib package provides an API to create and manage clusters and nodes. It uses an abstract interface called Driver to ensure that such can be created on multiple platforms.

To create a kutti client, one needs to reference this package along with one or more driver implementations, such as [github.com/kuttiproject/driver-vbox](https://github.com/kuttiproject/driver-vbox).

