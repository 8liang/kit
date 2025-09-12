package viperparser

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
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

func Unmarshal(rawVal any, opts ...Option) error {
	v, err := Parse(opts...)
	if err != nil {
		return err
	}
	return v.Unmarshal(rawVal)
}

func Parse(opts ...Option) (v *viper.Viper, err error) {
	p := New()
	p.Configure(opts...)
	if err = p.Parse(); err != nil {
		return
	}
	return p.viper, nil
}

func New() *Parser {
	return &Parser{
		viper:     viper.NewWithOptions(viper.ExperimentalBindStruct()),
		logger:    slog.Default(),
		envPrefix: "APP",
		envFile:   ".env",
	}
}

type Parser struct {
	viper     *viper.Viper
	logger    *slog.Logger
	url       string
	envPrefix string
	envFile   string
}

func (p *Parser) loadFromUrl(cfgAddr string) (err error) {
	var parsed *url.URL
	if parsed, err = url.Parse(cfgAddr); err != nil {
		return
	}
	p.viper.SetConfigType("yaml")
	switch parsed.Scheme {
	case SchemeFile:
		return p.loadFromFile(parsed.Path)
	case SchemeEtcdHttp, SchemeEtcdHttps:
		parsed.Scheme = strings.TrimLeft(parsed.Scheme, EtcdPrefix)
		return p.loadFromEtcd3(parsed)
	case SchemeHttp, SchemeHttps:
		return p.loadFromHttp(parsed)
	}
	return nil

}
func (p *Parser) Parse() (err error) {
	p.viper.SetEnvPrefix(p.envPrefix)
	p.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	p.viper.AutomaticEnv()
	// first load from the specified url
	if p.url != "" {
		if err = p.loadFromUrl(p.url); err != nil {
			return
		}
	}
	// then load from .env file
	return p.loadEnvFile()
}

func (p *Parser) Configure(opts ...Option) {
	for _, opt := range opts {
		opt(p)
	}
}
func (p *Parser) loadFromFile(filePath string) error {
	p.viper.SetConfigFile(filePath)
	return p.viper.ReadInConfig()
}

func (p *Parser) loadFromEtcd3(parsedUrl *url.URL) (err error) {
	endpoint := fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
	if err = p.viper.AddRemoteProvider("etcd3", endpoint, parsedUrl.Path); err != nil {
		return
	}
	err = p.viper.ReadRemoteConfig()
	return
}

func (p *Parser) loadFromHttp(parsedUrl *url.URL) (err error) {
	endpoint := fmt.Sprintf("%s://%s", parsedUrl.Scheme, parsedUrl.Host)
	var resp *req.Response
	if resp, err = req.R().Get(endpoint + parsedUrl.Path); err != nil {
		return fmt.Errorf("failed to load config from http: %w", err)
	}
	return p.viper.ReadConfig(resp.Body)
}

// loadEnvFile using viper to load .env file
func (p *Parser) loadEnvFile() (err error) {
	if _, err = os.Stat(p.envFile); err != nil {
		return nil
	}
	// create a new viper instance to read .env file
	envViper := viper.New()
	envViper.SetConfigFile(p.envFile)
	envViper.SetConfigType("dotenv")

	if err = envViper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to load config from env file: %w", err)
	}
	// foreach key in envViper, set it to p.viper if not exists in os.Getenv
	for _, key := range envViper.AllKeys() {
		envKey := strings.ToUpper(p.envPrefix + "_" + strings.ReplaceAll(key, ".", "_"))
		// only when the env variable does not have this key, use the .env file value
		if os.Getenv(envKey) == "" {
			p.viper.Set(key, envViper.Get(key))
		}
	}
	p.logger.Info("loaded .env file")
	return nil
}
