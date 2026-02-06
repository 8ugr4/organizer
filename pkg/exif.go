package pkg

import (
	"errors"
	"fmt"
	"github.com/barasher/go-exiftool"
	"os"
	"strings"
	"time"
)

var ErrNoCreateDate = errors.New("given file doesn't have a CreateDate field or we failed to find it")

func initExifTool() (*exiftool.Exiftool, error) {
	exifTool, err := exiftool.NewExiftool()
	if err != nil {
		return nil, err
	}

	return exifTool, nil
}

// getFileDate tries EXIF -> CreateDate and returns either month or year as string
// periodType is "month" or "year"
// if file doesn't have exif data return "" string
func (o *Operator) getFileDate(fp, periodType string) (string, error) { //nolint:unused
	f, err := os.Open(fp)
	if err != nil {
		return "", err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	var timePeriod string
	fileInfos := o.Storage.Exif.ExtractMetadata(f.Name())
	for _, fileInfo := range fileInfos {
		if fileInfo.Err != nil {
			return "", fileInfo.Err
		}
		if date, exists := fileInfo.Fields["CreateDate"]; exists {
			timePeriod = date.(string)
		}
	}

	if timePeriod != "" {
		parseTime, err := func(timePeriod, periodType string) (string, error) {
			ta, timeError := time.Parse("2006:01:02 15:04:05", timePeriod)
			if timeError != nil {
				return "", timeError
			}
			switch periodType {
			case "month":
				yearmonth := strings.Split(ta.Format("2006-01"), "-")

				return strings.Join(yearmonth, "/"), nil
			case "year":
				return ta.Format("2006"), nil
			default:
				return "", errors.New("no time thingy m8")
			}
		}(timePeriod, periodType)
		if err != nil {
			return "", err
		}
		if parseTime == "" {
			return "", fmt.Errorf("invalid periodType %s, must be 'month' or 'year'", periodType)
		}

		return parseTime, nil
	}

	return "", ErrNoCreateDate
}
