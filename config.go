package serverfinder

import "errors"

// Config - config for Find() func
type Config struct {
	PortStart int                  // Port from starting search a ProxyRotator
	PortEnd   int                  // Port end
	Request   func(port int) error // Text for send to ProxyRotator
}

// ------------------------------------------------------------------

func (c *Config) chk() error {
	if c == nil || c.Request == nil {
		return errors.New("c == nil || c.Request == nil")
	}
	return nil
}

// ------------------------------------------------------------------
