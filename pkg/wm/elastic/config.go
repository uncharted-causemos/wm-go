package elastic

// Config defines the parameters needed to instantiate a KB.
type Config struct {
	Addr string
}

// init fills in defaults for missing config parameters.
func (cfg *Config) init() {
	if cfg.Addr == "" {
		cfg.Addr = "http://localhost:9200"
	}
}
