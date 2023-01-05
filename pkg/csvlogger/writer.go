// Package csvlogger provides a simple CSV logger for Go.
// It is intented to store csv data in a file until the maximum file size is reached. Then a new file is created.
package csvlogger

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// SimulateFileSizeLimit can simulate a file too large error. Set it to a non-zero value to simulate a file size limit after n lines.
var SimulateFileSizeLimit = 0

// FileSizeLimitReached is returned if the file size limit is reached
type FileSizeLimitReached struct{}

func (m *FileSizeLimitReached) Error() string {
	return "file size limit reached"
}

// DiskFull is returned if the disk is full
type DiskFull struct{}

func (m *DiskFull) Error() string {
	return "disk full"
}

// Writer is a CSV logger
type Writer struct {
	Comma           rune // Comma is the field delimiter. It is set to ',' by NewWriter.
	outPath         string
	outFilePrefix   string
	writer          *csv.Writer
	currentFile     *os.File
	currentFileName string    // current file name with path
	lastFlush       time.Time // last flush time
	logger          zerolog.Logger
	lineCount       int
}

// NewWriter creates a new CSV logger.
func NewWriter(outPath string, outFilePrefix string) *Writer {
	return &Writer{
		Comma:           ',',
		outPath:         outPath,
		outFilePrefix:   outFilePrefix,
		writer:          nil,
		currentFile:     nil,
		currentFileName: "",
		lastFlush:       time.Time{},
		logger:          log.With().Str("component", "csvlogger").Logger(),
	}
}

// Write writes a single CSV record to w.
// If file size limit is reached, a FileSizeLimitReached error is returned. The current file is closed and a subsequent write will go into a new file.
// If disk is full, a DiskFull error is returned.
func (w *Writer) Write(record []string) error {
	if w.writer == nil {
		if err := w.newCsvWriter(); err != nil {
			return err
		}
	}

	err := w.writer.Write(record)
	if err != nil {
		err = w.handleWriteErrors(err)
		return fmt.Errorf("could not write record to file %s: %w", w.currentFileName, err)
	}
	w.lineCount++

	// simulate a file size limit
	if SimulateFileSizeLimit != 0 && w.lineCount > SimulateFileSizeLimit {
		err = w.handleWriteErrors(&os.PathError{
			Err: errors.New("file too large"),
		})
		return fmt.Errorf("simulated file size error %s: %w", w.currentFileName, err)
	}

	// check if its time to flush
	if time.Since(w.lastFlush) > 2*time.Second {
		w.writer.Flush()
		err := w.writer.Error()
		if err != nil {
			err = w.handleWriteErrors(err)
			return fmt.Errorf("could not flush file %s: %w", w.currentFileName, err)
		}
		w.lastFlush = time.Now()
	}
	return nil
}

func (w *Writer) handleWriteErrors(err error) error {
	var pathError *os.PathError

	if errors.As(err, &pathError) {
		if strings.Contains(pathError.Err.Error(), "file too large") {
			w.Close()
			w.logger.Warn().Msgf("file too large %s", w.currentFileName)
			return &FileSizeLimitReached{}
		}
		if strings.Contains(pathError.Err.Error(), "no space left on device") {
			w.logger.Warn().Msgf("disk full %s", w.currentFileName)
			return &DiskFull{}
		}
	}
	return err
}

func (w *Writer) newCsvWriter() error {
	// close current file
	w.Close()

	// create new file name
	fileName, err := w.nextFileName()
	if err != nil {
		return fmt.Errorf("could not create new file name: %w", err)
	}
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("could not create new file %s: %w", fileName, err)
	}
	w.currentFile = f
	w.currentFileName = fileName
	w.logger.Info().Msgf("created new file %s", fileName)
	w.writer = csv.NewWriter(f)
	w.writer.Comma = w.Comma
	w.lastFlush = time.Now()
	w.lineCount = 0
	return nil
}

// Close closes the Writer.
// subsequent writes to the Writer will go into a new file.
func (w *Writer) Close() {
	if w.writer != nil {
		w.writer.Flush()
		w.writer = nil
		if w.currentFile != nil {
			w.currentFile.Close()
			w.currentFile = nil
		}
	}
}

// scan the files in the output directory and find the next file name to use
func (w *Writer) nextFileName() (string, error) {
	// check what is the next file name to use
	files, err := os.ReadDir(w.outPath)
	if err != nil {
		return "", err
	}
	highestIndex := 0

	for _, file := range files {
		if strings.HasPrefix(file.Name(), w.outFilePrefix) {
			// get suffix
			s := strings.TrimPrefix(file.Name(), w.outFilePrefix)
			s = strings.TrimSuffix(s, ".csv")
			// convert to int
			i, err := strconv.Atoi(s)
			if err == nil {
				if i > highestIndex {
					highestIndex = i
				}
			}
		}
	}
	// create new file name
	return fmt.Sprintf("%s/%s%04d.csv", w.outPath, w.outFilePrefix, highestIndex+1), nil
}
