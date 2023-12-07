package listener

import (
    "context"
    "net"
    "os"
    "strings"
)

func isUnixDomainSocket(addr string) (bool, string) {
    if !strings.HasPrefix(addr, "unix:") {
        return false, ""
    }
    return true, strings.TrimPrefix(addr, "unix:")
}

func Listen(address string, ctx context.Context) (listener net.Listener, err error) {
    if ok, filename := isUnixDomainSocket(address); ok {
        return listenUnixDomainSocket(filename, ctx)
    }
    return net.Listen("tcp", address)
}

func listenUnixDomainSocket(filename string, ctx context.Context) (listener net.Listener, err error) {
    if listener, err = net.Listen("unix", filename); err != nil {
        panic(err)
    }
    if err = os.Chmod(filename, 0666); err != nil {
        panic(err)
    }
    go func() {
        <-ctx.Done()
        if err := os.Remove(filename); err != nil {
            panic(err)
        }
    }()
    return
}
