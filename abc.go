package abc

//Imports
import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

//Declarations
var (
	//Declare flags
	noDebug bool
	noProgress bool
	noKeep bool
	noDownload bool

	//Declare common variables
	start time.Time
	fileExists bool
	existingSize int64
	totalSizeStr string
	refTotalSize int64
	fileFlags int
	complete chan bool
	retry int
	response *http.Response
	oFile *os.File
	retryTimer time.Duration
)

//Display errors
func debug(err error) {
	if !noDebug {os.Stdout.WriteString("\n"+err.Error()+"\n")}
}

//Evaluate to retry or not
func canRetry(retryMax *int) (retryBool bool) {
	if retryMax != nil && *retryMax == retry {return}
	retry++
	return true
}

//Check if a download should resume
func canResume(acceptRanges string, byteRange *string) (rsm bool) {
	if existingSize > 0 &&
	acceptRanges == "bytes" &&
	byteRange == nil {return true}
	return
}

//Calculate and display progress
func printProgress(totalSize, currentSize, lastSize int64) {
	clearLine := "\r                                                                               "
	if totalSize > 0 {
		fmt.Printf(clearLine+"\r%.0f", (float64(currentSize)/float64(totalSize)*100))
	} else {
		os.Stdout.WriteString(clearLine+"\r--")
	}
	os.Stdout.WriteString("% | "+strconv.FormatInt(currentSize, 10)+" bytes of "+totalSizeStr+" | "+strconv.FormatInt((currentSize-lastSize), 10)+" bytes per second")
}

//Continue to refresh progress until download is complete
func progress(file *string, totalSize int64) {
	fileInfo, _ := os.Stat(*file)
	currentSize := fileInfo.Size()
	lastSize := currentSize
	for {
		select {
			case <-complete:
				printProgress(totalSize, currentSize, lastSize)
				complete <- true
				return
			default:
				printProgress(totalSize, currentSize, lastSize)
				time.Sleep(time.Second)
				lastSize = currentSize
				fileInfo, _ = os.Stat(*file)
				currentSize = fileInfo.Size()
		}
	}
}

//Signal and wait for go routine to print one last progress and terminate
func syncProgress() {
	if !noProgress {
		complete <- true
		<-complete
	}
}

func filePrep(file *string) (err error) {
	//Check if file already exists or not
	fileInfo, err := os.Stat(*file)
	if err == nil {
		if fileInfo.IsDir() {
			err = errors.New("A directory with that name already exists")
			debug(err)
			return
		}
		if noKeep && retry == 0 {
			err = os.Remove(*file)
			if err != nil {
				debug(err)
				return
			}
		} else {
			fileExists = true
			existingSize = fileInfo.Size()
		}
	} else {err = nil}
	//Create directory structure as needed
	os.MkdirAll(filepath.Dir(*file), 644)
	return
}

func requestHeaders(urlRaw *string, client *http.Client) (totalSize int64, acceptRanges string, err error) {
	//Request for response headers
	response, err = client.Head(*urlRaw)
	if err != nil {return}

	//Set size if available
	totalSizeStr = response.Header.Get("Content-Length")
	if totalSizeStr != "" {
		totalSize, err = strconv.ParseInt(totalSizeStr, 10, 64)
		if err == nil {
			if totalSize > 0 {
				if totalSize == existingSize {
					err = errors.New("The download was already completed previously")
					return
				} else if existingSize > totalSize {
					err = errors.New("Destination file larger than content length")
					return
				}
			}
		} else {totalSizeStr = "?"}
	} else {totalSizeStr = "?"}

	//Grab acceptRanges
	acceptRanges = response.Header.Get("Accept-Ranges")

	response.Body.Close()

	return
}

func retrySleep() {
	time.Sleep(retryTimer)
	if retryTimer.Minutes() < 5 {retryTimer *= 2}
}

//Public ABC Download function
func Download(urlRaw, file, byteRange, agent *string, timeout *time.Duration, retryMax, flags *int) (err error, totalSize int64, acceptRanges string) {

	//Conditional initializations
	if file == nil {noDownload = true}
	if flags != nil {
		if 1&*flags != 0 {noDebug = true}
		if 2&*flags != 0 {noProgress = true}
		if 4&*flags != 0 {noKeep = true}
	}

	//Exit now if no URL given
	if urlRaw == nil {
		err = errors.New("No URL provided")
		debug(err)
		return
	}

	//Set initial retry timer in milliseconds
	retryTimer = 200 * time.Millisecond

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
	if agent != nil {request.Header.Set("User-Agent",*agent)}

	for {
		if !noDownload {
			if !noDebug {
				if retry == 0 {
					os.Stdout.WriteString("Downloading "+*urlRaw+" to "+*file+"...\n")
				} else {
					os.Stdout.WriteString("Retry "+strconv.Itoa(retry)+"...\n")
					retrySleep()
				}
			}

			//Check and prepare file
			err = filePrep(file)
			if err != nil {
				debug(err)
				return
			}

			//Record start time
			if retry == 0 {start = time.Now()}
		}
		//Request for response headers
		totalSize, acceptRanges, err = requestHeaders(urlRaw, client)
		if err != nil {
			if err.Error() != "The download was already completed previously" && canRetry(retryMax) {continue}
			debug(err)
			return
		}
		break
	}

	//Remember original totalSize for reference in case it changes later
	refTotalSize = totalSize

	//Return now if noDownload
	if noDownload {
		err = nil
		return
	}

	//Set range header as needed if supplied by an argument
	if acceptRanges == "bytes" {
		if byteRange != nil {
			request.Header.Set("Range", "bytes="+*byteRange)
		}
	} else {
		if fileExists {
			err = os.Remove(*file)
			if err != nil {
				debug(err)
				return
			}
			fileExists = false
		}
	}

	firstTry := true
	for {
		if retry > 0 && !firstTry {
			if !noDebug {
				if !noProgress {os.Stdout.WriteString("\n")}
				os.Stdout.WriteString("Retry "+strconv.Itoa(retry)+"...\n")
			}

			oFile.Close()
			response.Body.Close()

			retrySleep()

			totalSize, acceptRanges, err = requestHeaders(urlRaw, client)
			if err != nil {
				debug(err)
				return
			}

			if refTotalSize != totalSize {
				err = errors.New("The Content-Length header has changed")
				debug(err)
				return
			}

			err = filePrep(file)
			if err != nil {
				debug(err)
				return
			}
		}

		firstTry = false

		//Set flags for creating/opening file
		if fileExists {
			if canResume(acceptRanges, byteRange) {
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
		oFile, err = os.OpenFile(*file, fileFlags, 644)
		if err != nil {
			debug(err)
			return
		}

		//If resuming, seek to the end of the existing file and set Range header to resume at next byte
		if canResume(acceptRanges, byteRange) {
			_, err = oFile.Seek(0, os.SEEK_END)
			if err != nil {
				debug(err)
				return
			}
			request.Header.Set("Range", "bytes="+strconv.FormatInt(existingSize, 10)+"-")
		}

		//Print progress as needed
		if !noProgress {

			//Initialize signaling channel
			complete = make(chan bool)

			//Start Go routine to print progress
			go progress(file, totalSize)
		}

		//Request for content
		response, err = client.Do(request)
		if err != nil {
			syncProgress()
			if canRetry(retryMax) {continue}
			debug(err)
			return
		}

		//Write content to file
		_, err = io.Copy(oFile, response.Body)
		if err != nil {
			syncProgress()
			if canRetry(retryMax) {continue}
			debug(err)
			return
		}

		// Close file and client response body
		oFile.Close()
		response.Body.Close()

		break
	}

	//Terminate progress if needed, display duration, and exit
	syncProgress()
	if !noDebug {os.Stdout.WriteString("\nDownload completed in "+time.Since(start).String()+"\n")}
	err = nil
	return
}