package main

import (
	"flag"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

func getCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "image-porter <CONFIG_FILE>",
		Short:             "an image syncing tool",
		DisableAutoGenTag: true,
		Args:              cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Help()
			}
			configFile := args[0]
			data, err := os.ReadFile(configFile)
			if err != nil {
				return err
			}
			var config Config
			err = yaml.Unmarshal(data, &config)
			if err != nil {
				return err
			}
			err = config.Init()
			if err != nil {
				return err
			}
			return Sync(&config)
		},
	}

	flags := cmd.PersistentFlags()
	initKlogFlags(flags)
	return cmd
}

// initKlogFlags 将 klog 需要的参数设置到全局
func initKlogFlags(flags *pflag.FlagSet) {
	klogFlags := flag.NewFlagSet("glog", flag.ExitOnError)
	klog.InitFlags(klogFlags)
	flags.AddGoFlagSet(klogFlags)
}
