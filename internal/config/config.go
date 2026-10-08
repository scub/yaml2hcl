// internal/config/config.go
package config

import "github.com/spf13/viper"


type Endpoint struct {
	Name		string 	`mapstructure:"name"`
	URL			string 	`mapstructure:"url"`
	Method		string 	`mapstructure:"method"`
	ExpectCode	int 	`mapstructure:"expect_code"`
	ExpectBody 	string 	`mapstructure:"expect_body"`
}

type Config struct {
	Timeout		int			`mapstructure:"timeout"`
	Verbose		bool		`mapstructure:"verbose"`
	Endpoints 	[]Endpoint	`mapstructure:"endpoints"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Apply defaults for endpoints
	for i := range cfg.Endpoints {
		if cfg.Endpoints[i].Method == "" {
			cfg.Endpoints[i].Method = "GET"
		}
		if cfg.Endpoints[i].ExpectCode == 0 {
			cfg.Endpoints[i].ExpectCode = 200
		}
	}

	return &cfg, nil
}