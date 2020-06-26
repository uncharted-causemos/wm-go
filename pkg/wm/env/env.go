package env

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Specification contains the imported environment variables.
type Specification struct {
	Addr string `default:":4200"`
	Mode string `default:"dev"`

	ElasticURL string `required:"true" envconfig:"ELASTIC_URL"`

	DgraphURLS []string `required:"true" envconfig:"DGRAPH_URLS"`

	MaasURL      string `required:"true" envconfig:"MAAS_URL"`
	MaasUser     string `required:"true" envconfig:"MAAS_USERNAME"`
	MaasPassword string `required:"true" envconfig:"MAAS_PASSWORD"`
}

// Load imports the environment variables and returns them in an Specification.
func Load(envFile string) (*Specification, error) {

	env := os.Getenv("WM_MODE")
	// if no env var in existing environment, load environment file from the .env file, otherwise (in production) just check existing host environment
	if "" == env {
		err := godotenv.Load(envFile)
		if err != nil {
			return nil, fmt.Errorf("Error loading %s file: %v", envFile, err)
		}
	}

	var s Specification
	err := envconfig.Process("wm", &s)
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
