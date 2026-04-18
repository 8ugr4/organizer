package pkg

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"
)

// CSVLogger writes log entries into a CSV file with three columns:
// sourceFileName, destinationFileName, SUCCESS/FAILURE.
type CSVLogger struct {
	mu     sync.Mutex
	writer *csv.Writer
	file   *os.File
}

// NewCSVLogger creates or truncates a CSV file and writes the header row.
func NewCSVLogger(path string) (*CSVLogger, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	w := csv.NewWriter(f)

	// header
	if err := w.Write([]string{"sourceFilePath", "destinationFilePath", "fileName", "status"}); err != nil {
		err2 := f.Close()

		return nil, fmt.Errorf("%w,%w", err, err2)
	}
	w.Flush()

	return &CSVLogger{writer: w, file: f}, nil
}

// Log writes single entry into the CSV file.
func (l *CSVLogger) Log(status, source, fileName, destination string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	record := []string{source, destination, fileName, status}
	if err := l.writer.Write(record); err != nil {
		return err
	}
	l.writer.Flush()

	return l.writer.Error()
}

func (l *CSVLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer.Flush()

	return l.file.Close()
}

func ResultLog(extensions int, o *Operator, startTime time.Time) {
	slog.Debug("", "unique extension count", extensions)
	slog.Debug("", "sub-dir count", o.SubDirCount)
	slog.Debug("", "skipped file count", len(o.Storage.Unprocessed))
	if len(o.Storage.Unprocessed) > 0 {
		for _, unprocessedFileName := range o.Storage.Unprocessed {
			slog.Warn("", "skipped", unprocessedFileName)
		}
	}
	slog.Info("", "total runtime", time.Since(startTime))
	if o.CsvHandler != nil {
		if err := o.CsvHandler.Log(time.Since(startTime).String(), "skipped file count", "total runtime", strconv.Itoa(len(o.Storage.Unprocessed))); err != nil {
			slog.Error("failure-log", "error", err.Error())
		}
	}
}
