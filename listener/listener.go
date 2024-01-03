package listener

import (
    "fmt"
    "net"
    "os"
    "strings"
)

func Listen(address string) (listener net.Listener, err error) {
    if ok, filename := isUnixDomainSocket(address); ok {
        return listenUnixDomain(filename)
    }
    return net.Listen("tcp", address)
}

func isUnixDomainSocket(addr string) (bool, string) {
    if !strings.HasPrefix(addr, "unix:") {
        return false, ""
    }
    return true, strings.TrimPrefix(addr, "unix:")
}

type domainSocketListener struct {
    net.Listener
    filename string
}

func (u *domainSocketListener) Accept() (net.Conn, error) {
    return u.Listener.Accept()
}

func (u *domainSocketListener) Close() error {
    defer func() {
        if err := os.Remove(u.filename); err != nil {
            fmt.Printf("remove unix domain socket file %q error %q", u.filename, err)
        }
    }()
    return u.Listener.Close()
}

func (u *domainSocketListener) Addr() net.Addr {
    return u.Listener.Addr()
}

func listenUnixDomain(filepath string) (_ net.Listener, err error) {
    l := &domainSocketListener{filename: filepath}
    if l.Listener, err = net.Listen("unix", filepath); err != nil {
        return
    }
    if err = os.Chmod(filepath, 0666); err != nil {
        _ = l.Close()
        return
    }
    return l, err
}
