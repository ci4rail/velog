package mvb

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/ci4rail/io4edge-client-go/functionblock"
	"github.com/ci4rail/io4edge-client-go/mvbsniffer"
	mvbpb "github.com/ci4rail/io4edge_api/mvbSniffer/go/mvbSniffer/v1"
	"github.com/ci4rail/velog/cmd/logger/internal/ctx"
	"github.com/ci4rail/velog/pkg/csvlogger"
	"github.com/ci4rail/velog/pkg/processdatastore"
)

// Run starts the MVB logger
func (l *Logger) Run() error {

	if l.cfg.DumpInterval < 10 {
		return fmt.Errorf("dump interval must be at least 10ms")
	}

	c, err := mvbsniffer.NewClientFromUniversalAddress(l.cfg.SnifferDevice, 0)
	if err != nil {
		return fmt.Errorf("error creating mvb sniffer client: %s", err)
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
		return fmt.Errorf("error starting mvb sniffer stream: %s", err)
	}

	s := processdatastore.NewStore()
	csvLogger := csvlogger.NewWriter(l.outputDir, l.cfg.FileName)
	writeCsvHeader(csvLogger)

	// go routine to read the stream and write it to the process data store
	go func() {
		l.logger.Info().Msg("Start logging MVB data")

		wg, err := ctx.WgFromContext(l.ctx)
		if err != nil {
			l.logger.Error().Msg(err.Error())
			return
		}
		defer wg.Done()

		for {
			select {
			case <-l.ctx.Done():
				l.logger.Info().Msg("Stop capture MVB data")
				return
			default:
			}
			sd, err := c.ReadStream(time.Second * 2)
			if err == nil {
				telegramCollection := sd.FSData.GetEntry()
				// l.logger.Info().Msgf("Read stream: %d", len(telegramCollection))

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

	// go routine to log the number of lines written to the csv file
	go func() {
		for {
			time.Sleep(5 * time.Second)
			l.logger.Info().Msgf("Number of lines written to all csv files: %d", l.lineCount)
		}
	}()

	return nil
}

func (l *Logger) logTelegram(s *processdatastore.Store, telegram *mvbpb.Telegram) {
	s.Write(newTelegramObject(telegram))
}

func (l *Logger) storeToCsv(s *processdatastore.Store, csvLogger *csvlogger.Writer) {
	defer csvLogger.Close()

	wg, err := ctx.WgFromContext(l.ctx)
	if err != nil {
		l.logger.Error().Msg(err.Error())
		return
	}
	defer wg.Done()

	for {
		time.Sleep(time.Duration(l.cfg.DumpInterval) * time.Millisecond)

		select {
		case <-l.ctx.Done():
			l.logger.Info().Msg("Stop storing MVB data")
			return
		default:
		}

		err := l.DumpStore(s, csvLogger, false, 0)
		l.dumpNumber++

		var diskFull *csvlogger.DiskFull
		if errors.As(err, &diskFull) {
			l.logger.Error().Msgf("Disk full when writing csv entry: %s. Stop recording", err)
			return
		}
	}
}

// DumpStore dumps the process data store to a csv file
// If dumpAll is true, all entries are dumped, otherwise only the entries that have been updated since the last dump
func (l *Logger) DumpStore(s *processdatastore.Store, csvLogger *csvlogger.Writer, dumpAll bool, recursionLevel int) error {
	addresses := s.List()
	for _, address := range addresses {
		o, updates, err := s.Read(uint32(address))
		if err == nil {
			if dumpAll || (updates > 0) {
				err := l.writeCsvEntry(csvLogger, o, updates)

				var fileSizeLimitReached *csvlogger.FileSizeLimitReached
				var diskFull *csvlogger.DiskFull

				if errors.As(err, &fileSizeLimitReached) {
					if recursionLevel > 0 {
						return fmt.Errorf("file size limit reached, but dumpStore was called recursively")
					}
					// a new file was created, write the header and dump the whole store
					writeCsvHeader(csvLogger)
					err := l.DumpStore(s, csvLogger, true, recursionLevel+1)

					if err != nil {
						l.logger.Error().Msgf("Error dumping store: %s", err)
					}
					return nil
				} else if errors.As(err, &diskFull) {
					return err
				} else if err != nil {
					l.logger.Error().Msgf("Error writing csv entry: %s", err)
				}
			}
		} else {
			l.logger.Error().Msgf("Error reading process data store: %s", err)
		}
	}
	return nil
}

func writeCsvHeader(csvLogger *csvlogger.Writer) {
	csvLogger.Write([]string{
		"Dump #",
		"Address (hex)",
		"Last Update - TimeSinceStart (us)",
		"Data (hex)",
		"FCode (dec)",
		"Updates (dec)",
		time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (l *Logger) writeCsvEntry(csvLogger *csvlogger.Writer, o processdatastore.Object, updates int) error {
	err := csvLogger.Write([]string{
		strconv.Itoa(l.dumpNumber),
		fmt.Sprintf("%x", o.Address()),
		fmt.Sprintf("%d", o.Timestamp()),
		hex.EncodeToString(o.Data()),
		o.AdditionalInfo()[0],
		strconv.Itoa(updates),
	})
	if err != nil {
		return err
	}
	l.lineCount++

	return nil
}
