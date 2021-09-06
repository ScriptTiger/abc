package main

import (
	"github.com/ScriptTiger/abc"
	"os"
	"strconv"
	"strings"
	"time"
)

//Function to display help text and exit
func help(err int) {
	os.Stdout.WriteString(
		"Usage: abc [options...] <url> <file>\n"+
		" -i <URL>            Source URL\n"+
		" -nodebug            Don't display debug messages, such as errors\n"+
		" -noprogress         Don't display progress\n"+
		" -o <file>           Destination file\n"+
		" -range <range>      Set Range header\n"+
		" -timeout <duration> Set connection timeout\n")
	os.Exit(err)
}

func main() {

	//Declare variables
	var err error

	//Declare and initialize flag pointers
	urlRaw := new(string)
	file := new(string)
	byteRange := new(string)
	timeout := new(time.Duration)
	flags := new(int)

	//Display help and exit if not enough arguments
	if len(os.Args) < 2 {help(-1)}

	//Push arguments to flag pointers
	for i := 1; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			switch strings.TrimPrefix(os.Args[i], "-") {
				case "range":
					i++
					if *byteRange == "" {byteRange = &os.Args[i]
					} else {help(-2)}
					continue
				case "i":
					i++
					if *urlRaw == "" {urlRaw = &os.Args[i]
					} else {help(-2)}
					continue
				case "timeout":
					i++
					if *timeout == time.Second*0 {
						*timeout, err = time.ParseDuration(os.Args[i])
						if err != nil {
							os.Stdout.WriteString(err.Error())
							os.Exit(-2)
						}
					} else {help(-2)}
					continue
				case "nodebug":
					if 1&*flags == 0 {*flags += 1
					} else {help(-2)}
					continue
				case "noprogress":
					if 2&*flags == 0 {*flags += 2
					} else {help(-2)}
					continue
				case "o":
					i++
					if *file == "" {file = &os.Args[i]
					} else {help(-2)}
					continue
				default:
					help(-2)
			}
		} else if *urlRaw == "" {urlRaw = &os.Args[i]
		} else if *file == "" {file = &os.Args[i]
		} else {help(-2)}
	}

	err, totalSize, acceptRanges := abc.Download(urlRaw, file, byteRange, timeout, flags)
	if err == nil {
		if *file == "" {
			os.Stdout.WriteString(
				"Content-Length = "+strconv.FormatInt(totalSize, 10)+
				"\nAccept-Ranges = "+acceptRanges)
		}
	} else {os.Exit(-4)}
	os.Exit(0)
}