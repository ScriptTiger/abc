[![Say Thanks!](https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg)](https://saythanks.io/to/thescripttiger%40gmail.com)

# ABC: Arbitrary Byte Collector  
ABC is a package for Go which can act as either a standard HTTP downloader with resume capabilities or as an arbitrary byte collector to download only arbitrary portions of files by manipulating the HTTP Range header as needed. ABC can easily be imported into any Go project and be implemented as Go routines to download arbitrary portions of files concurrently. The reference implementation of ABC launches a single instance of ABC, but can also be easily scripted for concurrency.

Usage: `abc [options...] <url> <file>`

Argument               | Description
-----------------------|--------------------------------------------------------------------------------------------------------
 `-i <URL>`            | Source URL
 `-nodebug`            | Don't display debug messages, such as errors
 `-noprogress`         | Don't display progress
 `-o <file>`           | Destination file
 `-range <range>`      | Set Range header
 `-timeout <duration>` | Set connection timeout

By default, if only a URL is given and no destination file, ABC will not download anything and just exit after returning any errors as well as the COntent-Length and Accept-Ranges headers. If both the URL and destination file are given, without any additional arguments, ABC will either create a new destination file if it doesn't exist already, append to the existing destination file if it does exist but is not complete, or delete the existing file if it does exist but the server does not support resume capabilities and create a new file. ABC gives you both enough information as well as flexibility to either easily implement your own additional functionality in Go or via a scripted solution using the included reference implementation.

# More About ScriptTiger

For more ScriptTiger scripts and goodies, check out ScriptTiger's GitHub Pages website:  
https://scripttiger.github.io/

[![Donate](https://www.paypalobjects.com/en_US/i/btn/btn_donateCC_LG.gif)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=MZ4FH4G5XHGZ4)

Donate Monero (XMR): 441LBeQpcSbC1kgangHYkW8Tzo8cunWvtVK4M6QYMcAjdkMmfwe8XzDJr1c4kbLLn3NuZKxzpLTVsgFd7Jh28qipR5rXAjx
