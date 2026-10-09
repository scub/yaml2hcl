// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string


var rootCmd = &cobra.Command{
	Use: "yaml2hcl",
	Short: "Convert yaml to HCL",
	Long: `yaml2hcl is a CLI tool that converts Yaml Helm values files into HCL,
allowing for quick conversion of prototypes to TF module format.`,
}


func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (defaults to $HOME/.healthcheck.yml)")
	rootCmd.PersistentFlags().IntP("timeout", "t", 5,
		"HTTP timeout in seconds")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false,
		"enable verbose output")

	if err := viper.BindPFlag("timeout", rootCmd.Flags().Lookup("timeout")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("verbose", rootCmd.Flags().Lookup("verbose")); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".healthcheck")
	}

	viper.SetEnvPrefix("HEALTHCHECK")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}