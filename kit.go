package kit

import (
	"net"

	"github.com/pkg/errors"
)

// IP 获取本机非环回IP地址
// IP gets the non-loopback IP address of the local machine
func IP() (string, error) {
	address, err := net.InterfaceAddrs()
	if err != nil {
		return "", errors.Wrap(err, "Get IP error")
	}
	for _, addr := range address {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", errors.New("Can not get IP")
}
