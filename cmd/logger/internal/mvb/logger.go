package mvb

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ci4rail/io4edge-client-go/functionblock"
	"github.com/ci4rail/io4edge-client-go/mvbsniffer"
	mvbpb "github.com/ci4rail/io4edge_api/mvbSniffer/go/mvbSniffer/v1"
	"github.com/ci4rail/mvb-can-logger/pkg/csvlogger"
	"github.com/ci4rail/mvb-can-logger/pkg/processdatastore"
)

// Run starts the MVB logger
func (l *Logger) Run() error {

	c, err := mvbsniffer.NewClientFromUniversalAddress(l.cfg.SnifferDevice, 0)
	if err != nil {
		l.logger.Error().Msgf("Error creating MVB sniffer client: %s", err)
		return err
	}
	// start stream
	err = c.StartStream(
		mvbsniffer.WithFilterMask(mvbsniffer.FilterMask{
			// receive any process data telegram, except timed out frames
			FCodeMask:             0x001F,
			Address:               0x0000,
			Mask:                  0x0000,
			IncludeTimedoutFrames: false,
		}),
		mvbsniffer.WithFBStreamOption(functionblock.WithBucketSamples(100)),
		mvbsniffer.WithFBStreamOption(functionblock.WithBufferedSamples(200)),
	)
	if err != nil {
		l.logger.Error().Msgf("Error starting MVB sniffer stream: %s", err)
		return err
	}

	s := processdatastore.NewStore()
	csvLogger := csvlogger.NewWriter(l.outputDir, l.cfg.FileName)
	writeCsvHeader(csvLogger)

	// go routine to read the stream and write it to the process data store
	go func() {
		for {
			sd, err := c.ReadStream(time.Second * 1)
			if err == nil {
				telegramCollection := sd.FSData.GetEntry()

				for _, telegram := range telegramCollection {

					if telegram.State != uint32(mvbpb.Telegram_kSuccessful) {
						if telegram.State&uint32(mvbpb.Telegram_kMissedMVBFrames) != 0 {
							l.logger.Warn().Msg("one or more MVB frames are lost in the device since the last telegram")
						}
						if telegram.State&uint32(mvbpb.Telegram_kMissedTelegrams) != 0 {
							l.logger.Warn().Msg("one or more telegrams are lost")
						}
					}
					l.logTelegram(s, telegram)
				}
			} else {
				l.logger.Warn().Msgf("Error reading MVB sniffer stream: %s", err)
			}
		}
	}()

	// write the process data store periodically to the csv file
	go l.storeToCsv(s, csvLogger)

	return nil
}

func (l *Logger) logTelegram(s *processdatastore.Store, telegram *mvbpb.Telegram) {
	s.Write(newTelegramObject(telegram))
}

func writeCsvHeader(csvLogger *csvlogger.Writer) {
	csvLogger.Write([]string{
		"Timestamp",
		"Address",
		"Data",
		"AdditionalInfo",
	})
}

func writeCsvEntry(csvLogger *csvlogger.Writer, o processdatastore.Object) error {
	return csvLogger.Write([]string{
		fmt.Sprintf("%d", o.Timestamp()),
		strconv.Itoa(int(o.Address())),
		hex.EncodeToString(o.Data()),
		strings.Join(o.AdditionalInfo(), ":"),
	})
}

func (l *Logger) storeToCsv(s *processdatastore.Store, csvLogger *csvlogger.Writer) {
	for {
		time.Sleep(time.Second * 1)
		addresses := s.List()
		for _, address := range addresses {
			o, updates, err := s.Read(address)
			if err == nil {
				if updates > 0 {
					err := writeCsvEntry(csvLogger, o)

					var fileSizeLimitReached *csvlogger.FileSizeLimitReached
					var diskFull *csvlogger.DiskFull

					if errors.As(err, &fileSizeLimitReached) {
						// a new file was created, write the header and the last entry again
						writeCsvHeader(csvLogger)
						err := writeCsvEntry(csvLogger, o)

						if err != nil {
							l.logger.Error().Msgf("Error writing csv entry: %s", err)
						}
					} else if errors.As(err, &diskFull) {
						l.logger.Error().Msgf("Disk full when writing csv entry: %s. Stop recording", err)
						return
					} else if err != nil {
						l.logger.Error().Msgf("Error writing csv entry: %s", err)
					}
				}
			}
		}
	}
}
