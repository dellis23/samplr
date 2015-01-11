package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"time"
)

var DEBUG bool

func extract_line(rs io.ReadSeeker) string {
	words := []byte{}
	char := make([]byte, 1)
	started := false
	for {
		rs.Read(char)

		if !started {
			if string(char) == "\n" {
				started = true
			}
			continue
		}

		if started {
			words = append(words, char[0])
			if string(char) == "\n" {
				return string(words)
			}
		}
	}
}

func log(message string, args ...interface{}) {
	if DEBUG {
		fmt.Printf(message, args)
	}
}

func main() {

	// Arguments
	var (
		filename string
		percent  float64
		number   int
	)
	flag.Float64Var(&percent, "p", 0, "percent of file to analyze")
	flag.IntVar(&number, "n", 0, "number of lines to analyze")
	flag.BoolVar(&DEBUG, "d", false, "whether to enable debug logging")
	flag.Parse()
	filename = flag.Args()[0]
	rand.Seed(time.Now().UTC().UnixNano())
	if percent == 0 && number == 0 {
		panic("Either percent or number must be passed.")
	}

	// Open / Setup File
	f, _ := os.Open(filename)
	defer f.Close()
	rs := io.ReadSeeker(f)
	file_info, err := f.Stat()
	if err != nil {
		panic(err)
	}
	file_size := file_info.Size()

	// Iterate and Print
	var (
		seek_size             int64
		current_seek_location int64

		// ... for percentage calculation
		bytes_read int64

		// ... for line-number calculation
		lines_read          int
		average_line_length int
	)
	for {

		average_line_length = 32

		// ... skip to new location
		if percent > 0 {
			seek_size = int64(float64(1) / (percent / float64(100)) * float64(average_line_length) * 2)
		} else {
			lines_left_in_file := (file_size - current_seek_location) / int64(average_line_length)
			lines_still_needed := number - lines_read
			seek_size = int64(float64(1) / (float64(lines_still_needed) /
				float64(lines_left_in_file)) * float64(average_line_length) * 2)
		}
		log("seek_size: %d\n", seek_size)
		randomized_seek_size := rand.Int63n(seek_size)
		log("randomized_seek_size: %d\n", randomized_seek_size)
		current_seek_location, err = rs.Seek(randomized_seek_size, 1)
		if err != nil {
			panic("Could not seek to location in file!")
		}
		log("current_seek_location: %d\n", current_seek_location)

		text := extract_line(rs)
		fmt.Print(text)

		// ... track progress
		if percent > 0 {
			bytes_read += int64(len(text))
		} else {
			if average_line_length > 0 {
				average_line_length = ((average_line_length * lines_read) + len(text)) / (lines_read + 1)
			} else {
				average_line_length = len(text)
			}
			lines_read++
		}

		// ... stop if done
		if current_seek_location+seek_size > file_size {
			break
		}
	}
}
