package abc

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func Download(urlRaw, file, byteRange *string, timeout *time.Duration, flags *int) (err error, totalSize int64, acceptRanges string) {

	//Declarations
	var (
		//Declare flags
		noDebug bool
		noProgress bool
		noDownload bool

		//Declare common variables
		fileInfo fs.FileInfo
		start time.Time
		existingSize int64
		totalSizeStr string
		fileExists bool
		fileFlags int
		complete chan bool
	)

	//Conditional initializations
	if file == nil {noDownload = true}
	if flags != nil {
		if 1&*flags != 0 {noDebug = true}
		if 2&*flags != 0 {noProgress = true}
	}

	//Function to handle errors
	debug := func(err error) {
		if !noDebug {os.Stdout.WriteString("\n"+err.Error()+"\n")}
	}

	//Exit now if no URL given
	if urlRaw == nil {
		err = errors.New("No URL provided\n")
		debug(err)
		return
	}

	if !noDownload {


		if !noDebug {os.Stdout.WriteString("Downloading "+*urlRaw+" to "+*file+"...\n")}

		//Check if file already exists or not
		fileInfo, err = os.Stat(*file)
		if err == nil {
			if fileInfo.IsDir() {
				err = errors.New("A directory with that name already exists\n")
				debug(err)
				return
			}
			fileExists = true
			existingSize = fileInfo.Size()
		}
		//Create directory structure as needed
		os.MkdirAll(filepath.Dir(*file), 644)
	}

	//Initialize HTTP client
	if timeout == nil {timeout = new(time.Duration)}
	client := &http.Client{Timeout: *timeout}

	//Initialize HTTP request
	request, err := http.NewRequest("GET", *urlRaw, nil)
	if err != nil {
		debug(err)
		return
	}

	//Set HTTP request headers
	request.Header.Set("Accept-Language","en-US")
	request.Header.Set("Connection","Keep-Alive")
	request.Header.Set("User-Agent","Mozilla/5.0")

	if !noDownload {
		//Record start time
		start = time.Now()
	}

	//Request for response headers
	headers, err := client.Head(*urlRaw)
	if err != nil {
		debug(err)
		return
	}
	defer headers.Body.Close()

	//Set size if available
	totalSizeStr = headers.Header.Get("Content-Length")
	if totalSizeStr != "" {
		totalSize, err = strconv.ParseInt(totalSizeStr, 10, 64)
		if err != nil {
			if totalSize > 0 && totalSize == existingSize {
				if !noDebug {os.Stdout.WriteString("The download was already completed previously\n")}
				err = nil
				return
			}
		}
	} else {totalSizeStr = "?"}

	//Grab acceptRanges
	acceptRanges = headers.Header.Get("Accept-Ranges")

	//Return now if noDownload
	if noDownload {
		err = nil
		return
	}

	//Set range header as needed if supplied by an argument
	if acceptRanges == "bytes" && byteRange != nil {
		request.Header.Set("Range", "bytes="+*byteRange)
	}

	//Function to check if a download should resume
	canResume := func() (rsm bool) {
		if existingSize > 0 &&
		acceptRanges == "bytes" &&
		byteRange == nil {
			rsm = true
		} else {rsm = false}
		return
	}

	//Set flags for creating/opening file
	if fileExists {
		if canResume() {
			fileFlags = os.O_APPEND | os.O_WRONLY
		} else {
			err = os.Remove(*file)
			if err != nil {
				debug(err)
				return
			}
			fileExists = false
			fileFlags = os.O_CREATE | os.O_WRONLY
		}
	} else {fileFlags = os.O_CREATE | os.O_WRONLY}

	//Initialize oFile
	oFile, err := os.OpenFile(*file, fileFlags, 644)
	if err != nil {
		debug(err)
		return
	}

	//If resuming, seek to the end of the existing file and set Range header to resume at next byte
	if canResume() {
		_, err = oFile.Seek(0, os.SEEK_END)
		if err != nil {
			debug(err)
			return
		}
		defer oFile.Close()
		request.Header.Set("Range", "bytes="+strconv.FormatInt(existingSize, 10)+"-")
	}

	//Print progress as needed
	if !noProgress {

		//Initialize signaling channel
		complete = make(chan bool)

		//Go routine to print progress
		go func() {
			clearLine := "\r                                                                               "
			fileInfo, _ = os.Stat(*file)
			currentSize := fileInfo.Size()
			lastSize := currentSize
			printProgress := func() {
				if totalSize > 0 {
					fmt.Printf(clearLine+"\r%.0f", (float64(currentSize)/float64(totalSize)*100))
				} else {
					os.Stdout.WriteString(clearLine+"\r--")
				}
				os.Stdout.WriteString("% | "+strconv.FormatInt(currentSize, 10)+" bytes of "+totalSizeStr+" | "+strconv.FormatInt((currentSize-lastSize), 10)+" bytes per second")
			}
			for {
				select {
					case <-complete:
						printProgress()
						complete <- true
						return
					default:
						printProgress()
						time.Sleep(time.Second)
						lastSize = currentSize
						fileInfo, _ = os.Stat(*file)
						currentSize = fileInfo.Size()
				}
			}
		}()
	}

	//Request for content
	response, err := client.Do(request)
	if err != nil {
		debug(err)
		return
	}
	defer response.Body.Close()

	//Write content to file
	_, err = io.Copy(oFile, response.Body)
	if err != nil {
		debug(err)
		return
	}

	//Signal and wait for go routine to print last progress and terminate
	if !noProgress {
		complete <- true
		<-complete
	}

	//Display duration and exit
	if !noDebug {os.Stdout.WriteString("\nDownload completed in "+time.Since(start).String())}
	err = nil
	return
}