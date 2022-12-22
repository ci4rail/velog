package can

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ci4rail/io4edge-client-go/canl2"
	"github.com/ci4rail/io4edge-client-go/functionblock"
	canpb "github.com/ci4rail/io4edge_api/canL2/go/canL2/v1alpha1"
	"github.com/ci4rail/velog/cmd/logger/internal/ctx"
	"github.com/ci4rail/velog/pkg/csvlogger"
)

// Run starts the CAN logger
func (l *Logger) Run() error {

	c, err := canl2.NewClientFromUniversalAddress(l.cfg.SnifferDevice, 0)
	if err != nil {
		return fmt.Errorf("error creating can sniffer client: %s", err)
	}

	err = c.UploadConfiguration(
		canl2.WithBitRate(uint32(l.cfg.Bitrate)),
		canl2.WithSamplePoint(l.cfg.SamplePoint),
		canl2.WithSJW(uint8(l.cfg.SJW)),
		canl2.WithListenOnly(true),
	)
	if err != nil {
		return fmt.Errorf("error uploading can sniffer configuration: %s", err)
	}

	// start stream
	err = c.StartStream(
		canl2.WithFilter(l.cfg.AcceptanceCode, l.cfg.AcceptanceMask),
		canl2.WithFBStreamOption(functionblock.WithBucketSamples(50)),
		canl2.WithFBStreamOption(functionblock.WithBufferedSamples(100)),
	)
	if err != nil {
		return fmt.Errorf("error starting can sniffer stream: %s", err)
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
				l.logger.Info().Msg("Stop logging CAN data")
				return
			default:
			}
			sd, err := c.ReadStream(time.Second * 2)
			if err == nil {
				samples := sd.FSData.Samples
				//l.logger.Info().Msgf("Read CAN sniffer stream: %d", len(samples))

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
		}
	}()

	// go routine to log abnormal controller state to journal
	go func() {
		wg, err := ctx.WgFromContext(l.ctx)
		if err != nil {
			l.logger.Error().Msg(err.Error())
			return
		}
		defer wg.Done()
		for {
			select {
			case <-l.ctx.Done():
				l.logger.Info().Msg("Stop logging CAN controller state")
				return
			default:
			}
			// check for abnormal controller state
			state, err := c.GetCtrlState()
			if err != nil {
				l.logger.Warn().Msgf("Error getting CAN sniffer controller state: %s", err)
			}

			if canpb.ControllerState(state) != canpb.ControllerState_CAN_OK {
				l.logger.Warn().Msgf("CAN sniffer controller state: %s", canpb.ControllerState(state).Enum().String())
			}
			time.Sleep(2 * time.Second)
		}
	}()

	// go routine to log the number of lines written to the csv file to journal
	go func() {
		for {
			time.Sleep(5 * time.Second)
			l.logger.Info().Msgf("Number of lines written to all csv files: %d", l.lineCount)
		}
	}()
	return nil
}

func (l *Logger) Write(s *canpb.Sample, csvLogger *csvlogger.Writer) error {
	err := l.writeCsvEntry(csvLogger, s)

	var fileSizeLimitReached *csvlogger.FileSizeLimitReached
	var diskFull *csvlogger.DiskFull

	if errors.As(err, &fileSizeLimitReached) {
		// a new file was created, write the header and the last entry again
		writeCsvHeader(csvLogger)
		err := l.writeCsvEntry(csvLogger, s)

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

func (l *Logger) writeCsvEntry(csvLogger *csvlogger.Writer, s *canpb.Sample) error {
	rtr := ""
	if s.Frame.RemoteFrame {
		rtr = "R"
	}
	ext := ""
	if s.Frame.ExtendedFrameFormat {
		ext = "X"
	}
	err := csvLogger.Write([]string{
		fmt.Sprintf("%d", s.Timestamp),
		fmt.Sprintf("%x", s.Frame.MessageId),
		hex.EncodeToString(s.Frame.Data),
		ext,
		rtr,
	})
	if err != nil {
		return err
	}
	l.lineCount++
	return nil
}
