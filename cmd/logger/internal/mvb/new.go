package mvb

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type configuration struct {
	SnifferDevice string // e.g. "S101-IOU03-USB-EXT-1-mvbSniffer"
	FileName      string // prefix for log files e.g. "mvb"
	DumpInterval  int    // how often to dump the store to csv file in ms
}

// Logger is the instance of the MVB logger
type Logger struct {
	cfg       *configuration
	outputDir string
	logger    zerolog.Logger
	ctx       context.Context
	lineCount int
}

// NewFromViper creates a new MVB Unit from a viper configuration
func NewFromViper(ctx context.Context, viperCfg *viper.Viper, outputDir string) (*Logger, error) {
	cfg, err := readConfig(viperCfg)
	if err != nil {
		return nil, err
	}
	return New(ctx, cfg, outputDir), nil
}

// New creates a new instance of MVB Unit
func New(ctx context.Context, cfg *configuration, outputDir string) *Logger {

	l := &Logger{
		cfg:       cfg,
		outputDir: outputDir,
		logger:    log.With().Str("component", "MVB").Logger(),
		ctx:       ctx,
		lineCount: 0,
	}

	l.logger.Info().Msg(fmt.Sprintf("config: %+v", cfg))

	return l
}

func readConfig(sub *viper.Viper) (*configuration, error) {
	if sub == nil {
		return nil, fmt.Errorf("missing configuration")
	}
	var cfg configuration
	err := sub.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config %s", err)
	}

	return &cfg, nil
}
