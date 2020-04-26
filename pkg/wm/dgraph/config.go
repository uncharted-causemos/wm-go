package dgraph

// Config defines the parameters needed to instantiate a DB.
type Config struct {
	Addrs []string
}

// init fills in defaults for missing config parameters.
func (cfg *Config) init() {
	if len(cfg.Addrs) == 0 {
		cfg.Addrs = []string{"localhost:9080"}
	}
}
