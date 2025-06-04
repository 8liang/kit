package viperparser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

const (
	SchemeFile      = ""
	EtcdPrefix      = "etcd+"
	SchemeHttp      = "http"
	SchemeHttps     = "https"
	SchemeEtcdHttp  = EtcdPrefix + SchemeHttp
	SchemeEtcdHttps = EtcdPrefix + SchemeHttps
)

func Parse(cfgAddr string) (err error) {
	var parsed *url.URL
	if parsed, err = url.Parse(cfgAddr); err != nil {
		return
	}
	switch parsed.Scheme {
	case SchemeFile:
		err = loadFromFile(parsed.Path)
	case SchemeEtcdHttp, SchemeEtcdHttps:
		parsed.Scheme = strings.TrimLeft(parsed.Scheme, EtcdPrefix)
		err = loadFromEtcd3(parsed)
	case SchemeHttp, SchemeHttps:
		err = loadFromHttp(parsed)
	}
	return
}
func loadFromFile(filePath string) error {
	viper.SetConfigFile(filePath)
	return viper.ReadInConfig()
}

func loadFromEtcd3(parsedUrl *url.URL) (err error) {
	endpoint := fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
	if err = viper.AddRemoteProvider("etcd3", endpoint, parsedUrl.Path); err != nil {
		return
	}
	err = viper.ReadRemoteConfig()
	return
}

func loadFromHttp(parsedUrl *url.URL) (err error) {
	endpoint := fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
	var resp *req.Response
	if resp, err = req.R().Get(endpoint + parsedUrl.Path); err != nil {
		return fmt.Errorf("failed to load config from http: %w", err)
	}
	return viper.ReadConfig(resp.Body)
}
