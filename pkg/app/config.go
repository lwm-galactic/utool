package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

const configFlagName = "config"

// a parameter reception config flag parse.
var cfgFile string

func init() {
	pflag.StringVarP(&cfgFile, "config", "c", cfgFile, "Read configuration from specified `FILE`, "+
		"support JSON, TOML, YAML, HCL, or Java properties formats.")
}

// addConfigFlag adds flags for a specific server to the specified FlagSet object.
func addConfigFlag(commandName string, fs *pflag.FlagSet) {
	fs.AddFlag(pflag.Lookup(configFlagName))

	// env
	viper.AutomaticEnv()
	viper.SetEnvPrefix(strings.Replace(strings.ToUpper(commandName), "-", "_", -1))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	cobra.OnInitialize(func() {
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else { // if no set --config flag
			// search for now dir
			viper.AddConfigPath(".")

			if names := strings.Split(commandName, "-"); len(names) > 1 {
				// TODO add search path homedir
				// viper.AddConfigPath(filepath.Join(homedir.HomeDir(), "."+names[0]))
				viper.AddConfigPath(filepath.Join("/etc", names[0]))
			}

			viper.SetConfigName(commandName)
		}

		if err := viper.ReadInConfig(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: failed to read configuration file(%s): %v\n", cfgFile, err)
			os.Exit(1)
		}
	})
}
