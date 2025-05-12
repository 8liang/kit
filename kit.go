package kit

import (
	"fmt"
	"net"
	"os"
	"strconv"

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

// ParseAddr 解析地址
// ParseAddr parses the address
func ParseAddr(addr string) (host string, port int, err error) {
	var p string
	host, p, err = net.SplitHostPort(addr)
	if err != nil {
		return
	}

	port, err = strconv.Atoi(p)
	if err != nil {
		return
	}
	return
}

// IsInDocker 检测是否在Docker中运行
// IsInDocker checks if the application is running in Docker
func IsInDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return false
}

func IsInK8s() bool {
	// 检查特定的环境变量
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		return true
	} else if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount"); err == nil {
		return true
	}
	return false
}

func GetHostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		data, errr := os.ReadFile("/etc/hostname")
		if errr != nil {
			return "", fmt.Errorf("get hostname error, %v, %w", err, errr)
		}
		hostname = string(data)
	}
	return hostname, nil
}
