# gRPC Retry Golang Library

![CI](https://github.com/kulti/grpc-retry/workflows/CI/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/kulti/grpc-retry)](https://goreportcard.com/report/github.com/kulti/grpc-retry)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/kulti/grpc-retry.svg)](https://pkg.go.dev/github.com/kulti/grpc-retry)

gRPC golang library does not allow to reconnect after failures. This package provides a server wrapper to handle that.

## Example

```
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()

srv := server.New(server.Params{
    Address:      "0.0.0.0:9090",
    RetryTimeout: time.Second,
    OnRetryFn:    func(err error) { log.Println("restarting server after error:", err) },
})

go func() {
    <-ctx.Done()
    srv.GracefulStop()
}()

myrpc.RegisterMyServer(srv, myrpcserver.New())
srv.Run(ctx)
```
