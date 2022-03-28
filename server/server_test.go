package server_test

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/kulti/grpc-retry/server"
)

func TestListenFailed(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	requireNoError(t, err)

	retryCh := make(chan error)

	params := server.Params{
		Address:      l.Addr().String(),
		RetryTimeout: time.Millisecond,
		OnRetryFn: func(err error) {
			retryCh <- err
		},
	}

	srv := server.New(params)
	requireServerNotReady(t, srv)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	go func() {
		defer close(retryCh)
		srv.Run(ctx)
	}()

	retryErr := waitRetryError(t, retryCh)
	requireAddressAlreadyInUseError(t, retryErr)

	requireNoError(t, l.Close())

	unlockRetryErrors(retryCh)

	waitServerReady(t, srv)
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func requireServerNotReady(t *testing.T, srv *server.Server) {
	t.Helper()
	if srv.Ready() {
		t.Fatalf("server should not be ready here")
	}
}

func waitRetryError(t *testing.T, retryCh <-chan error) error {
	t.Helper()

	var retryErr error
	requireEventually(t, func() bool {
		select {
		case retryErr = <-retryCh:
			return true
		default:
			return false
		}
	}, time.Second, time.Millisecond)

	return retryErr
}

func unlockRetryErrors(retryCh <-chan error) {
	go func() {
		for range retryCh {
		}
	}()
}

func waitServerReady(t *testing.T, srv *server.Server) {
	t.Helper()
	requireEventually(t, srv.Ready, time.Second, time.Millisecond)
}

func requireEventually(t *testing.T, condition func() bool, waitFor time.Duration, tick time.Duration) {
	t.Helper()

	ch := make(chan bool, 1)

	timer := time.NewTimer(waitFor)
	defer timer.Stop()

	ticker := time.NewTicker(tick)
	defer ticker.Stop()

	for tick := ticker.C; ; {
		select {
		case <-timer.C:
			t.Fatal("Condition never satisfied")
		case <-tick:
			tick = nil
			go func() { ch <- condition() }()
		case v := <-ch:
			if v {
				return
			}
			tick = ticker.C
		}
	}
}

func requireAddressAlreadyInUseError(t *testing.T, err error) {
	if !isAddressAlreadyInUseError(err) {
		t.Fatalf("error should be 'address already in used', but: %v", err)
	}
}

func isAddressAlreadyInUseError(err error) bool {
	var errErrno syscall.Errno
	if !errors.As(err, &errErrno) {
		return false
	}

	return errErrno == syscall.EADDRINUSE
}
