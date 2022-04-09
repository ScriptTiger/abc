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
		" -agent <user-agent> Set User-Agent header\n"+
		" -i <URL>            Source URL\n"+
		" -retry <number>     Number of retries\n"+
		" -nodebug            Don't display debug messages, such as errors\n"+
		" -nokeep             If file already exists, delete and download new\n"+
		" -noprogress         Don't display progress\n"+
		" -o <file>           Destination file\n"+
		" -range <range>      Set Range header\n"+
		" -timeout <duration> Set connection timeout\n")
	os.Exit(err)
}

func main() {

	//Declarations
	var (
		urlRaw *string
		file *string
		byteRange *string
		agent *string
		timeout *time.Duration
		retry *int
		flags *int
		err error
	)

	//Display help and exit if not enough arguments
	if len(os.Args) < 2 {help(-1)}

	//Push arguments to flag pointers
	for i := 1; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			switch strings.TrimPrefix(os.Args[i], "-") {
				case "range":
					i++
					if byteRange == nil {byteRange = &os.Args[i]
					} else {help(-2)}
					continue
				case "agent":
					i++
					if agent == nil {agent = &os.Args[i]
					} else {help(-2)}
					continue
				case "i":
					i++
					if urlRaw == nil {urlRaw = &os.Args[i]
					} else {help(-2)}
					continue
				case "timeout":
					i++
					if timeout == nil {
						timeout = new(time.Duration)
						*timeout, err = time.ParseDuration(os.Args[i])
						if err != nil {
							os.Stdout.WriteString(err.Error())
							os.Exit(-2)
						}
					} else {help(-2)}
					continue
				case "nodebug":
					if flags == nil {flags = new(int)}
					if 1&*flags == 0 {*flags += 1
					} else {help(-2)}
					continue
				case "noprogress":
					if flags == nil {flags = new(int)}
					if 2&*flags == 0 {*flags += 2
					} else {help(-2)}
					continue
				case "nokeep":
					if flags == nil {flags = new(int)}
					if 4&*flags == 0 {*flags += 4
					} else {help(-2)}
					continue
				case "o":
					i++
					if file == nil {file = &os.Args[i]
					} else {help(-2)}
					continue
				case "retry":
					i++
					if retry == nil {
						retry = new(int)
						*retry, err = strconv.Atoi(os.Args[i])
						if err != nil {help(-2)}
					} else {help(-2)}
					continue
				default:
					help(-2)
			}
		} else if urlRaw == nil {urlRaw = &os.Args[i]
		} else if file == nil {file = &os.Args[i]
		} else {help(-2)}
	}

	err, totalSize, acceptRanges := abc.Download(urlRaw, file, byteRange, agent, timeout, retry, flags)
	if err == nil {
		if file == nil {
			os.Stdout.WriteString(
				"Content-Length = "+strconv.FormatInt(totalSize, 10)+
				"\nAccept-Ranges = "+acceptRanges)
		}
	} else {os.Exit(1)}
	os.Exit(0)
}