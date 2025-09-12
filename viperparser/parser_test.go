package viperparser

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ParserTest struct {
	suite.Suite
}

func TestParser(t *testing.T) {
	suite.Run(t, new(ParserTest))
}

type Stub struct {
	Version    string `mapstructure:"version"`
	PublicRoot string `mapstructure:"public_root"`
	Web        struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"web"`
}

func (p *ParserTest) TestUnmarshal() {
	p.Run("load from .env", func() {
		cfg := &Stub{}
		err := Unmarshal(&cfg)
		p.Nil(err)
		p.Equal("v1.0.0", cfg.Version)
		p.Equal("127.0.0.1:80", cfg.Web.Addr)
	})
	p.Run("load from system env", func() {
		_ = os.Setenv("APP_VERSION", "v1.5.0")
		_ = os.Setenv("APP_PUBLIC_ROOT", "/public_root")
		_ = os.Setenv("APP_WEB_ADDR", "192.168.0.200:8080")
		cfg := &Stub{}
		err := Unmarshal(&cfg)
		p.Nil(err)
		p.Equal("v1.5.0", cfg.Version)
		p.Equal("192.168.0.200:8080", cfg.Web.Addr)
		p.Equal("/public_root", cfg.PublicRoot)
	})

	p.Run("load from config file", func() {
		cfg := &Stub{}
		err := Unmarshal(&cfg, WithUrl("./config.yaml"), WithEnvPrefix("IGNORE_"), WithEnvFile(".env.ignore"))
		p.Nil(err)
		p.Equal("v2.0.5", cfg.Version)
		p.Equal("./public", cfg.PublicRoot)
		p.Equal("localhost:8080", cfg.Web.Addr)
	})
}
