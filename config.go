package serverfinder

import (
	"errors"
)

// Config - config for Find() func
type Config struct {
	EnvVar    string                                       // Environment variable of server
	PortStart int                                          // Port from starting search a server
	PortEnd   int                                          // Port end
	Request   func(port int) (proxyPort int, err error) // Request func for send to server
}

// ------------------------------------------------------------------

func (c *Config) chk() error {
	if c == nil || c.Request == nil {
		return errors.New("c == nil || c.Request == nil")
	}
	return nil
}

// ------------------------------------------------------------------
