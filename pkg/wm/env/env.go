package env

import (
	"encoding/json"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Specification contains the imported environment variables.
type Specification struct {
	Addr string `default:":4200"`
	Mode string `default:"dev"`

	ElasticURL string `required:"true" envconfig:"ELASTIC_URL"`

	DgraphURLS []string `required:"true" envconfig:"DGRAPH_URLS"`
}

// Load imports the environment variables and returns them in an Specification.
func Load(envFile string) (*Specification, error) {
	err := godotenv.Load(envFile)
	if err != nil {
		return nil, fmt.Errorf("Error loading %s file: %v", envFile, err)
	}

	var s Specification
	err = envconfig.Process("wm", &s)
	if err != nil {
		return nil, fmt.Errorf("Error processing environment config: %v", err)
	}

	settings, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal env: %v", err)
	}
	fmt.Printf("Environment Settings:\n%s\n", string(settings))

	return &s, err
}
