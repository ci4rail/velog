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
	"os/signal"
	"syscall"

	"github.com/ci4rail/velog/cmd/logger/internal/can"
	"github.com/ci4rail/velog/cmd/logger/internal/ctx"
	"github.com/ci4rail/velog/cmd/logger/internal/mvb"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type globalConfiguration struct {
	LoggerOutputDir string
}

var (
	cfgFile   string
	globalCfg globalConfiguration
)

var rootCmd = &cobra.Command{
	Use:   "velog",
	Short: "Logs MVB and CAN data to disk",
	Long:  `Logs MVB and CAN data to disk`,
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	ctx, cancel, wg := ctx.NewWgContext()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05.999Z07:00"})

	err := viper.Unmarshal(&globalCfg)
	if err != nil {
		log.Fatal().Msgf("unmarshal global config %s", err)
	}

	// create logger output dir
	err = os.MkdirAll(globalCfg.LoggerOutputDir, 0755)
	if err != nil {
		log.Fatal().Msgf("create logger output dir %s", err)
	}

	// configure loggers
	var mvbLogger *mvb.Logger
	mvbConfig := viper.Sub("mvb")
	if mvbConfig != nil {
		mvbLogger, err = mvb.NewFromViper(ctx, mvbConfig, globalCfg.LoggerOutputDir)
		if err != nil {
			log.Fatal().Msgf("mvbLogger: %s", err)
		}
	}
	var canLogger *can.Logger
	canConfig := viper.Sub("can")
	if canConfig != nil {
		canLogger, err = can.NewFromViper(ctx, canConfig, globalCfg.LoggerOutputDir)
		if err != nil {
			log.Fatal().Msgf("canLogger: %s", err)
		}
	}

	// start loggers
	if mvbLogger != nil {
		mvbLogger.Run()
	}
	if canLogger != nil {
		canLogger.Run()
	}

	// Wait for termination signal
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, syscall.SIGTERM, syscall.SIGINT)
	sig := <-cancelChan
	log.Info().Msgf("Received signal %s", sig)
	cancel()
	wg.Wait()
	log.Info().Msgf("Exit Program")
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Msgf("Execute Root cmd: %s", err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".velog-config.yaml", "config file")
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
