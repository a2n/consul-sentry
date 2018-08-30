# 💂 Consul Sentry [![GoDoc][doc-img]][doc]

## Getting Started
Consul sentry is a Go package, that receives consul watch event, and notifiies caller by function or channel.

## Prerequisites
According to Consul agent [watch](https://www.consul.io/docs/agent/watches.html) documentation, the configuration must having explicit header `type`, supported types:

*  `key`
*  `keyprefix`
*  `services`
*  `nodes`
*  `service`
*  `checks`
*  `event`

## Example
Please referring [`test/server.go`](test/server.go)

[doc-img]: https://godoc.org/github.com/a2n/consul-sentry?status.svg
[doc]: https://godoc.org/github.com/a2n/consul-sentry