// cmd/check.go
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/scub/yaml2hcl/internal/checker"
	"github.com/scub/yaml2hcl/internal/config"
)

var checkCmd = &cobra.Command{
	Use: "check",
	Short: "Check all configured endpoints",
	Long: `Send requests to all configured endpoints and report their status back`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		if len(cfg.Endpoints) == 0 {
			return fmt.Errorf("no endpoints configured; create $HOME/.healthcheck.yaml file or use --config")
		}

		timeout := time.Duration(cfg.Timeout) * time.Second
		c := checker.New(timeout, cfg.Verbose)

		endpoints := make([]checker.Endpoint, len(cfg.Endpoints))
		for i, ep := range cfg.Endpoints {
			endpoints[i] = checker.Endpoint{
				Name: 			ep.Name,
				URL:			ep.URL,
				Method: 		ep.Method,
				ExpectCode:		ep.ExpectCode,
				ExpectBody: 	ep.ExpectBody,
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout+5*time.Second)
		defer cancel()

		results := c.CheckAll(ctx, endpoints)

		// Print Results
		unhealthy := 0
		for _, r := range results {
			fmt.Println(r.Summary())
			if !r.Healthy {
				unhealthy++
			}
		}

		fmt.Printf("\n%d/%d endpoints healthy\n", len(results)-unhealthy, len(results))

		if unhealthy > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringSliceP("url", "u", nil,
		"check a single URL (can be repeated)")

	viper.BindPFlag("urls", checkCmd.Flags().Lookup("url"))
}