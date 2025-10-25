// This file is part of https://github.com/MortenHarding/rss3270cli/
// Copyright 2025 by Morten Harding, licensed under the MIT license. See
// LICENSE in the project root for license information.

// It is based on example5 of https://github.com/racingmars/go3270/
// Copyright 2025 by Matthew R. Wilson
// and the code in https://github.com/ErnieTech101/rss3270svr
// Copyright ErnieTech101

package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/racingmars/go3270"
)

func rssurl(conn net.Conn, devinfo go3270.DevInfo, rssFeedURL any) (
	go3270.Tx, any, error) {

	currentURL := rssFeedURL.(string)

	// Accept Enter; PF3 exit.
	pfkeys := []go3270.AID{go3270.AIDEnter, go3270.AIDPF2, go3270.AIDPF3}
	exitkeys := []go3270.AID{go3270.AIDPF9}

	// Make a local copy of the screen definition that we can append lines to.
	screen := make(go3270.Screen, len(layout))
	copy(screen, layout)

	title := "Change channel"
	header := padCenter(title, 79)

	//Header
	screen = append(screen,
		go3270.Field{Row: 0, Col: 0, Content: header, Color: go3270.White, Intense: true},
		go3270.Field{Row: 1, Col: 0, Content: strings.Repeat("-", 79), Color: go3270.Blue}, // ASCII only
		go3270.Field{Row: 2, Col: 0, Content: "Enter URL:"},
		go3270.Field{Row: 2, Col: 11, Name: "newURL", Write: true, Highlighting: go3270.Underscore},
		go3270.Field{Row: 2, Col: 79, Autoskip: true}, // field "stop" character
		go3270.Field{Row: 3, Col: 0, Content: "Or select from one of the below channels:"},
		go3270.Field{Row: 3, Col: 42, Write: true, Name: "choice", Content: "0", Color: go3270.Turquoise},
	)

	// Build list of RSS Url's
	row := 4

	for i, url := range rssFeeds {
		for _, line := range wrap80(fmt.Sprintf("%2d. %s", i, url), 80) {
			if row >= 22 { // leave space for footer/input
				break
			}
			screen = append(screen, go3270.Field{Row: row, Col: 0, Content: line, Color: go3270.Yellow})
			row++
		}
		if row >= 22 {
			break
		}
	}

	//Footer
	screen = append(screen,
		go3270.Field{Row: 22, Col: 0, Content: strings.Repeat("-", 80), Color: go3270.Blue}, // ASCII only
		go3270.Field{Row: 23, Col: 0, Content: "Enter", Color: go3270.Turquoise},
		go3270.Field{Row: 23, Col: 6, Content: "Save & return", Color: go3270.Blue},
		go3270.Field{Row: 23, Col: 22, Content: "F2", Color: go3270.Turquoise},
		go3270.Field{Row: 23, Col: 25, Content: "Headlines", Color: go3270.Blue},
		go3270.Field{Row: 23, Col: 45, Content: "F3", Color: go3270.Turquoise},
		go3270.Field{Row: 23, Col: 48, Content: "Return", Color: go3270.Blue},
		go3270.Field{Row: 23, Col: 69, Content: "F9", Color: go3270.Turquoise},
		go3270.Field{Row: 23, Col: 72, Content: "Exit", Color: go3270.Blue},
	)

	fieldValues := make(map[string]string)

	// We can call the old HandleScreen(), or we could have used the new
	// HandleScreenAlt() and provided a nil DevInfo.
	resp, err := go3270.HandleScreen(
		screen,      // the screen to display
		nil,         // (no) rules to enforce
		fieldValues, // pre-populated values in fields
		pfkeys,      // keys we accept -- validating
		exitkeys,    // keys we accept -- non-validating
		"errormsg",  // name of field to put error messages in
		3, 43,       // cursor coordinates
		conn, // network connection
	)
	if err != nil {
		return nil, nil, err
	}

	switch resp.AID {
	case go3270.AIDEnter:
		fieldValues = resp.Values
		if fieldValues["choice"] != "" {
			ch := fieldValues["choice"]
			var i int
			if _, err := fmt.Sscanf(ch, "%2d", &i); err == nil {
				//Do something with the error
			}
			if i < len(rssFeeds) {
				currentURL = rssFeeds[i]
			} else {
				currentURL = rssFeeds[0]
			}

		}
		if fieldValues["newURL"] != "" {
			currentURL = fieldValues["newURL"]
		}
		// Save and go back
		return rssfeed, currentURL, nil
	case go3270.AIDPF2:
		// switch to Title screen
		return rsstitles, currentURL, nil
	case go3270.AIDPF3:
		// Exit
		return rssfeed, currentURL, nil
	case go3270.AIDPF9:
		// Exit
		return nil, nil, nil
	default:
		// re-run current transaction
		return rssfeed, currentURL, nil
	}
}
