/*
Copyright © 2022 Ci4Rail GmbH <engineering@ci4rail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type globalConfiguration struct {
	NatsAddress string // nats address. If set, publish directly to nats. Otherwise, publish to dapr pubsub
	NetworkName string // edgefarm network name, required for dapr pubsub
}

var (
	cfgFile   string
	globalCfg globalConfiguration
)

var rootCmd = &cobra.Command{
	Use:   "mvb-can-logger",
	Short: "Logs MVB and CAN data to disk",
	Long:  `Logs MVB and CAN data to disk`,
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05.999Z07:00"})

	err := viper.Unmarshal(&globalCfg)
	if err != nil {
		log.Fatal().Msgf("unmarshal global config %s", err)
	}
	select {}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Msgf("Execute Root cmd: %s", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".mvb-can-logger-config.yaml", "config file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName(cfgFile) // name of config file (without extension)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/") // call multiple times to add many search paths
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		log.Fatal().Msgf("fatal error config file: %s", err)
	}
}
