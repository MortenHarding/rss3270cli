This file is part of https://github.com/MortenHarding/rss3270cli/

Copyright 2025 by Morten Harding, licensed under the MIT license.

It is based on example5 of https://github.com/racingmars/go3270/

Copyright 2025 by Matthew R. Wilson

and the code in https://github.com/ErnieTech101/rss3270svr

Copyright ErnieTech101

# A simple RSS proxy for TN3270 emulators

This is a proxy server for use with **3270 (TN3270)** emulators, that displays an RSS feed on a 24×80 style “green screen”, using the `racingmars/go3270` library.

---
## Features

- Connect via a 3270 emulator (e.g. `x3270`, `c3270`, Vista or Mocha for Mac) to port **7300**  
- Customize the port used
- Displays top headlines from a selected RSS feed  
- Switch between different RSS feeds
- Add a custom RSS feed
- Customize the list of RSS feeds presented using the file `rssfeed.url`
- First row in `rssfeed.url` is the default RSS feed
- Handle some special characters, not in EBCDIC. Currently only Nordic characters.
- Refresh the RSS feed when you press **Enter**
- Select another RSS feed by pressing **PF4**   

---
## Requirements

- Network access from client to rss3270cli on port 7300, which is the default, or to the port defined by using the command line parameter -port xxxx
- The file [rssfeed.url](https://github.com/MortenHarding/rss3270cli/blob/main/rssfeed.url)
- A TN3270 emulator on client side

---
## How to use it

Get the latest releae of [rss3270cli](https://github.com/MortenHarding/rss3270cli/releases) from github, and the file [rssfeed.url](https://github.com/MortenHarding/rss3270cli/blob/main/rssfeed.url). Place both files in the same directory, and start rss3270cli.

 `./rss3270cli`

The default port is 7300 that you will access from your TN3270 terminal emulator.
Select your own port, using the command line parameter -port

 `./rss3270cli -port 9010`

---
## How to connect

Connect to the server's IP with a 3270 Client using port 7300

Example: `c3270 localhost:7300`

---
## Compile your own rss3270cli executable

 `git clone https://github.com/MortenHarding/rss3270cli.git`

 `cd rss3270cli`
 
 `go mod init rss3270cli`

Add the github racingmars/Go3270 dependency:
   
 `go get github.com/racingmars/go3270@latest`
 
 `go mod tidy`

Build an executable

 `go build -o rss3270cli .`
 

---
## License / Attribution

This library is licensed under the MIT license; see the file LICENSE for details.
