package can

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type configuration struct {
	SnifferDevice  string  // e.g. "S101-IOU03-USB-EXT-1-can"
	FileName       string  // prefix for log files e.g. "can"
	Bitrate        int     // e.g. 500000
	SamplePoint    float32 // e.g. 0.8
	SJW            int     // e.g. 1
	AcceptanceMask uint32  // e.g. 0x000
	AcceptanceCode uint32  // e.g. 0x7FF
}

// Logger is the instance of the CAN logger
type Logger struct {
	cfg       *configuration
	outputDir string
	logger    zerolog.Logger
	ctx       context.Context
}

// NewFromViper creates a new CAN Unit from a viper configuration
func NewFromViper(ctx context.Context, viperCfg *viper.Viper, outputDir string) (*Logger, error) {
	cfg, err := readConfig(viperCfg)
	if err != nil {
		return nil, err
	}
	return New(ctx, cfg, outputDir), nil
}

// New creates a new instance of CAN Unit
func New(ctx context.Context, cfg *configuration, outputDir string) *Logger {

	l := &Logger{
		cfg:       cfg,
		outputDir: outputDir,
		logger:    log.With().Str("component", "CAN").Logger(),
		ctx:       ctx,
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
