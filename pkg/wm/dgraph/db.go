package dgraph

import (
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
)

// DB wraps the client and serves as the basis of the graph.DB interface.
type DB struct {
	client *dgo.Dgraph
}

// New instantiates and returns a new DB using the provided Config.
func New(cfg *Config) (*DB, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.init()

	dialOpts := append([]grpc.DialOption{},
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))
	cs := make([]api.DgraphClient, len(cfg.Addrs))
	for i, addr := range cfg.Addrs {
		d, err := grpc.Dial(addr, dialOpts...)
		if err != nil {
			return nil, err
		}
		cs[i] = api.NewDgraphClient(d)
	}

	return &DB{
		client: dgo.NewDgraphClient(cs...),
	}, nil
}
