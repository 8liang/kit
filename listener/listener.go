package listener

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Listen 监听指定地址，支持Unix域套接字和TCP
// Listen listens on the specified address, supporting both Unix domain socket and TCP
func Listen(address string) (listener net.Listener, err error) {
	if ok, filename := isUnixDomainSocket(address); ok {
		return listenUnixDomain(filename)
	}
	return net.Listen("tcp", address)
}

// isUnixDomainSocket 检查地址是否为Unix域套接字
// isUnixDomainSocket checks if the address is a Unix domain socket
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

// listenUnixDomain 监听Unix域套接字
// listenUnixDomain listens on a Unix domain socket
func listenUnixDomain(sockAddr string) (_ net.Listener, err error) {
	l := &domainSocketListener{filename: sockAddr}
	var addr *net.UnixAddr
	if addr, err = net.ResolveUnixAddr("unix", sockAddr); err != nil {
		return
	}
	if l.Listener, err = net.ListenUnix("unix", addr); err != nil {
		return
	}
	if err = os.Chmod(sockAddr, 0666); err != nil {
		_ = l.Close()
		return
	}
	return l, err
}
