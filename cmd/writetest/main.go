package main

import (
	"fmt"
	"os"

	"time"

	"errors"
	"reflect"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// cat /sys/block/mmcblk1/device/cid
// 0353445343313647801b27202f014b00, see https://www.cameramemoryspeed.com/sd-memory-card-faq/reading-sd-card-cid-serial-psn-internal-numbers/
// write speed: approx 5 MByte/s on vFAT with this write test

// if the file size limit is reached, os.PathError is with error = "file too large" is thrown
// if no space is left on the device, os.PathError is with error = "no space left on device" is thrown

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05.999Z07:00"})

	log.Info().Msg("Hello World")

	// open file for writing
	f, err := os.Create("/mnt/test3.txt")
	//f, err := os.OpenFile("/mnt/test.txt",os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal().Msgf("create file: %s", err)
	}

	for {
		t := time.Now()
		_, err := f.WriteString(fmt.Sprintf("Hello World: %v\n", t))
		if err != nil {
			var pathError *os.PathError
			if errors.As(err, &pathError) {
				fmt.Printf("pathError: %+v\n", pathError)
			}
			log.Fatal().Msgf("write to file: %+v %s", err, reflect.TypeOf(err))
			// get PathError
			// write to file: write /mnt/test.txt: no space left on device *os.PathError

		}
		//time.Sleep(3 * time.Millisecond)
	}
}
