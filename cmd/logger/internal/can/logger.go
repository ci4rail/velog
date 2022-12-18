package can

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ci4rail/io4edge-client-go/canl2"
	"github.com/ci4rail/io4edge-client-go/functionblock"
	canpb "github.com/ci4rail/io4edge_api/canL2/go/canL2/v1alpha1"
	"github.com/ci4rail/mvb-can-logger/cmd/logger/internal/ctx"
	"github.com/ci4rail/mvb-can-logger/pkg/csvlogger"
)

// Run starts the CAN logger
func (l *Logger) Run() error {

	c, err := canl2.NewClientFromUniversalAddress(l.cfg.SnifferDevice, 0)
	if err != nil {
		l.logger.Error().Msgf("Error creating CAN sniffer client: %s", err)
		return err
	}

	err = c.UploadConfiguration(
		canl2.WithBitRate(uint32(l.cfg.Bitrate)),
		canl2.WithSamplePoint(l.cfg.SamplePoint),
		canl2.WithSJW(uint8(l.cfg.SJW)),
		canl2.WithListenOnly(true),
	)
	if err != nil {
		l.logger.Error().Msgf("Error uploading CAN sniffer configuration: %s", err)
		return err
	}

	// start stream
	err = c.StartStream(
		canl2.WithFilter(l.cfg.AcceptanceCode, l.cfg.AcceptanceMask),
		canl2.WithFBStreamOption(functionblock.WithBucketSamples(50)),
		canl2.WithFBStreamOption(functionblock.WithBufferedSamples(100)),
	)
	if err != nil {
		l.logger.Error().Msgf("Error starting CAN sniffer stream: %s", err)
		return err
	}

	csvLogger := csvlogger.NewWriter(l.outputDir, l.cfg.FileName)
	writeCsvHeader(csvLogger)

	// go routine to read the stream and write it to the csv file
	go func() {
		l.logger.Info().Msg("Start logging CAN data")
		defer csvLogger.Close()

		wg, err := ctx.WgFromContext(l.ctx)
		if err != nil {
			l.logger.Error().Msg(err.Error())
			return
		}
		defer wg.Done()

		for {
			select {
			case <-l.ctx.Done():
				return
			default:
			}
			sd, err := c.ReadStream(time.Second * 2)
			// l.logger.Info().Msgf("Read stream %v", err)

			if err == nil {
				samples := sd.FSData.Samples
				for _, sample := range samples {
					if sample.IsDataFrame {
						err := l.Write(sample, csvLogger)
						if err != nil {
							return
						}
					}
					if sample.Error != 0 {
						l.logger.Warn().Msgf("CAN sniffer stream error: %s", sample.Error.Enum().String())
					}
				}
			} else {
				l.logger.Warn().Msgf("Error reading CAN sniffer stream: %s", err)
			}

			// check for abnormal controller state
			state, err := c.GetCtrlState()
			if err != nil {
				l.logger.Warn().Msgf("Error getting CAN sniffer controller state: %s", err)
			}

			if canpb.ControllerState(state) != canpb.ControllerState_CAN_OK {
				l.logger.Warn().Msgf("CAN sniffer controller state: %s", canpb.ControllerState(state).Enum().String())
			}
		}
	}()

	return nil
}

func writeCsvHeader(csvLogger *csvlogger.Writer) {
	csvLogger.Write([]string{
		"TimeSinceStart (us)",
		"ID (hex)",
		"Data (hex)",
		"Ext",
		"RTR",
		time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (l *Logger) Write(s *canpb.Sample, csvLogger *csvlogger.Writer) error {
	err := writeCsvEntry(csvLogger, s)

	var fileSizeLimitReached *csvlogger.FileSizeLimitReached
	var diskFull *csvlogger.DiskFull

	if errors.As(err, &fileSizeLimitReached) {
		// a new file was created, write the header and the last entry again
		writeCsvHeader(csvLogger)
		err := writeCsvEntry(csvLogger, s)

		if err != nil {
			l.logger.Error().Msgf("Error writing csv entry: %s", err)
		}
	} else if errors.As(err, &diskFull) {
		l.logger.Error().Msgf("Disk full when writing csv entry: %s. Stop recording", err)
		return err
	} else if err != nil {
		l.logger.Error().Msgf("Error writing csv entry: %s", err)
	}
	return nil
}

func writeCsvEntry(csvLogger *csvlogger.Writer, s *canpb.Sample) error {
	rtr := ""
	if s.Frame.RemoteFrame {
		rtr = "R"
	}
	ext := ""
	if s.Frame.ExtendedFrameFormat {
		ext = "X"
	}
	return csvLogger.Write([]string{
		fmt.Sprintf("%d", s.Timestamp),
		fmt.Sprintf("%x", s.Frame.MessageId),
		hex.EncodeToString(s.Frame.Data),
		ext,
		rtr,
	})
}
