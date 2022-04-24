package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vitwit/resolute/client"
	"gopkg.in/yaml.v2"
)

type Config struct {
	DefaultChain string                               `yaml:"default_chain" json:"default_chain"`
	Chains       map[string]*client.ChainClientConfig `yaml:"chains" json:"chains"`

	cl map[string]*client.ChainClient
}

// createConfig creates config
func createConfig(home string, debug bool) error {
	cfgPath := path.Join(home, "config.yaml")

	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		if _, err := os.Stat(home); os.IsNotExist(err) {
			if err = os.Mkdir(home, os.ModePerm); err != nil {
				return err
			}
		}
	}

	f, err := os.Create(cfgPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write(defaultConfig(path.Join(home, "keys"), debug)); err != nil {
		return err
	}
	return nil
}

func overwriteConfig(cfg *Config) error {
	home := viper.GetString("home")
	cfgPath := path.Join(home, "config.yaml")
	f, err := os.Create(cfgPath)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(cfg.MustYAML()); err != nil {
		return err
	}

	log.Printf("updated resolute configuration at %s", cfgPath)
	return nil
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cmd *cobra.Command) error {
	home, err := cmd.PersistentFlags().GetString(flags.FlagHome)
	if err != nil {
		return err
	}

	debug, err := cmd.Flags().GetBool("debug")
	if err != nil {
		return err
	}
	config = &Config{}
	cfgPath := path.Join(home, "config.yaml")
	_, err = os.Stat(cfgPath)
	if err != nil {
		err = createConfig(home, debug)
		if err != nil {
			return err
		}
	}

	viper.SetConfigFile(cfgPath)
	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println("Failed to read in config:", err)
		os.Exit(1)
	}

	// read the config file bytes
	file, err := ioutil.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	// unmarshall them into the struct
	if err = yaml.Unmarshal(file, config); err != nil {
		fmt.Println("Error unmarshalling config:", err)
		os.Exit(1)
	}

	// instantiate chain client
	// TODO: this is a bit of a hack, we should probably have a
	// better way to inject modules into the client
	config.cl = make(map[string]*client.ChainClient)
	for name, chain := range config.Chains {
		chain.Modules = append([]module.AppModuleBasic{}, ModuleBasics...)
		cl, err := client.NewChainClient(chain, home, os.Stdin, os.Stdout)
		if err != nil {
			fmt.Println("Error creating chain client:", err)
			os.Exit(1)
		}
		config.cl[name] = cl
	}

	if cmd.PersistentFlags().Changed("output") {
		output, err := cmd.PersistentFlags().GetString("output")
		if err != nil {
			return err
		}

		// Should output be a global configuration item?
		for chain, _ := range config.Chains {
			config.Chains[chain].OutputFormat = output
		}
	}

	// validate configuration
	if err = validateConfig(config); err != nil {
		fmt.Println("Error parsing chain config:", err)
		os.Exit(1)
	}

	return nil
}

func defaultConfig(keyHome string, debug bool) []byte {
	return Config{
		DefaultChain: "cosmoshub",
		Chains: map[string]*client.ChainClientConfig{
			"cosmoshub": client.GetCosmosHubConfig(keyHome, debug),
		},
	}.MustYAML()
}

// MustYAML returns the yaml string representation of the Paths
func (c Config) MustYAML() []byte {
	out, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return out
}

func validateConfig(c *Config) error {
	for _, chain := range c.Chains {
		if err := chain.Validate(); err != nil {
			return err
		}
	}
	if c.GetDefaultClient() == nil {
		return fmt.Errorf("default chain (%s) configuration not found", c.DefaultChain)
	}
	return nil
}
func (c *Config) GetDefaultClient() *client.ChainClient {
	return c.GetClient(c.DefaultChain)
}

func (c *Config) GetClient(chainID string) *client.ChainClient {
	if v, ok := c.cl[chainID]; ok {
		return v
	}
	return nil
}
