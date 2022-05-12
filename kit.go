package kit

import (
    "github.com/pkg/errors"
    "net"
)

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
