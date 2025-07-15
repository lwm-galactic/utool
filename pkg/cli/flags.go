package cli

import (
	"flag"
	"github.com/lwm-galactic/logger"
	"github.com/spf13/pflag"
	"strings"
)

// WordSepNormalizeFunc changes all flags that contain "_" separators.
func WordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
	}
	return pflag.NormalizedName(name)
}

// WarnWordSepNormalizeFunc changes and warns for flags that contain "_" separators.
func WarnWordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		newName := strings.ReplaceAll(name, "_", "-")
		logger.Warnf("%s is DEPRECATED and will be removed in a future version. Use %s instead.", name, newName)

		return pflag.NormalizedName(newName)
	}
	return pflag.NormalizedName(name)
}

// InitFlags normalizes, parses, then logs the command line flags.
func InitFlags(flags *pflag.FlagSet) {
	flags.SetNormalizeFunc(WordSepNormalizeFunc)
	flags.AddGoFlagSet(flag.CommandLine)
}

// PrintFlags logs the flags in the flagset.
func PrintFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		logger.Debugf("FLAG: --%s=%q", flag.Name, flag.Value)
	})
}
