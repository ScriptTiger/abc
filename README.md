[![Say Thanks!](https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg)](https://docs.google.com/forms/d/e/1FAIpQLSfBEe5B_zo69OBk19l3hzvBmz3cOV6ol1ufjh0ER1q3-xd2Rg/viewform)

# ABC: Arbitrary Byte Collector  
ABC is a package for Go which can act as either a standard HTTP sequential downloader supporting streaming playback while downloading for supported "Web" videos (with supported players, such as VLC, MPlayer, ffplay, MPC-HC, etc.) and resume capabilities or as an arbitrary byte collector to download only arbitrary portions of files by manipulating the HTTP Range header as needed. ABC can easily be imported into any Go project and be implemented as Go routines to download arbitrary portions of files concurrently. The reference implementation of ABC launches a single instance of ABC, but can also be easily scripted for concurrency.

To import ABC into your project:  
`go get github.com/ScriptTiger/abc`  
Then just `import "github.com/ScriptTiger/abc"` and call abc.Download(...) to use.

Please refer to the dev package docs and reference implementation for more details and ideas on how to integrate ABC into your project.

Dev package docs:  
https://pkg.go.dev/github.com/ScriptTiger/abc

Reference implementation:  
https://github.com/ScriptTiger/abc/blob/main/ref/ref.go

# Reference Implementation

Usage: `abc [options...] <url> <file>`

Argument               | Description
-----------------------|--------------------------------------------------------------------------------------------------------
 `-agent <user-agent>` | Set User-Agent header
 `-i <URL>`            | Source URL
 `-retry <number>`     | Number of retries
 `-nodebug`            | Don't display debug messages, such as errors
 `-nokeep`             | If file already exists, delete and download new
 `-noprogress`         | Don't display progress
 `-o <file>`           | Destination file
 `-range <range>`      | Set Range header
 `-timeout <duration>` | Set connection timeout

By default, if only a URL is given and no destination file, ABC will not download anything and just exit after returning any errors as well as the Content-Length and Accept-Ranges headers. If both the URL and destination file are given, without any additional arguments, ABC will either create a new destination file if it doesn't exist already, append to the existing destination file if it does exist but is not complete, or delete the existing file if it does exist but the server does not support resume capabilities and create a new file. ABC gives you both enough information as well as flexibility to either easily implement your own additional functionality in Go or via a scripted solution using the included reference implementation.

# More About ScriptTiger

For more ScriptTiger scripts and goodies, check out ScriptTiger's GitHub Pages website:  
https://scripttiger.github.io/
