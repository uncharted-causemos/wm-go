package elastic

// ES wraps the client and serves as the basis of the wm.KnowledgeBase interface.
type ES struct {
	// client *someElasticsearchClient
}

// New instantiates and returns a new KB using the provided Config.
func New(cfg *Config) (*ES, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.init()

	// Connect to Elasticsearch here...
	return &ES{
		// client: ...
	}, nil
}
