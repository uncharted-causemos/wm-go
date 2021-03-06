package elastic

import (
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
	"gitlab.uncharted.software/WM/wm-go/pkg/wm"
)

// ES wraps the client and serves as the basis of the wm.KnowledgeBase interface.
type ES struct {
	client *elasticsearch.Client
}

// New instantiates and returns a new KB using the provided Config.
func New(cfg *Config) (*ES, error) {
	op := "ES.New"
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.init()

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{
			cfg.Addr,
		},
	})
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}

	res, err := client.Info()
	if err != nil {
		return nil, &wm.Error{Op: op, Err: err}
	}
	defer res.Body.Close()
	fmt.Printf("ES Client:\n%v\n", res)

	return &ES{client}, nil
}
