package config

type CheckConfig struct {
    Enabled   bool
    Func      func() (bool, string)
    Threshold float64
}

type Configuration struct {
    Checks map[string]CheckConfig
}

func NewConfiguration() *Configuration {
    return &Configuration{
        Checks: make(map[string]CheckConfig),
    }
}

func (c *Configuration) AddCheck(name string, check func() (bool, string)) {
    c.Checks[name] = CheckConfig{
        Enabled: true,
        Func:    check,
    }
}

func (c *Configuration) EnableCheck(name string) {
    if check, exists := c.Checks[name]; exists {
        check.Enabled = true
        c.Checks[name] = check
    }
}

func (c *Configuration) DisableCheck(name string) {
    if check, exists := c.Checks[name]; exists {
        check.Enabled = false
        c.Checks[name] = check
    }
}