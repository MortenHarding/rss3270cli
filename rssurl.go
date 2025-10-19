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

var newURL = rssFeeds[0]

func rssurl(conn net.Conn, devinfo go3270.DevInfo, data any) (
	go3270.Tx, any, error) {

	// Accept Enter; PF3 exit.
	pfkeys := []go3270.AID{go3270.AIDEnter}
	exitkeys := []go3270.AID{go3270.AIDPF3}

	// Make a local copy of the screen definition that we can append lines to.
	screen := make(go3270.Screen, len(layout))
	copy(screen, layout)

	title := "Change RSS URL Feed"
	header := padCenter(title, 79)

	//Header
	screen = append(screen,
		go3270.Field{Row: 0, Col: 0, Content: header, Intense: true},
		go3270.Field{Row: 1, Col: 0, Content: strings.Repeat("-", 79)}, // ASCII only
		go3270.Field{Row: 2, Col: 0, Content: "Enter URL: "},
		go3270.Field{Row: 2, Col: 10, Name: "newURL", Write: true, Highlighting: go3270.Underscore},
		go3270.Field{Row: 2, Col: 79, Autoskip: true}, // field "stop" character
		go3270.Field{Row: 3, Col: 0, Content: " or select from one of the below URL's"},
	)

	// Build list of RSS Url's
	row := 5

	for i, url := range rssFeeds {
		for _, line := range wrap80(fmt.Sprintf("%2d. %s", i, url), 80) {
			if row >= 20 { // leave space for footer/input
				break
			}
			screen = append(screen, go3270.Field{Row: row, Col: 0, Content: line})
			row++
		}
		if row >= 20 {
			break
		}
	}

	//Footer
	screen = append(screen,
		go3270.Field{Row: 20, Col: 0, Intense: true, Color: go3270.Red, Name: "errormsg"}, // a blank field for error messages
		go3270.Field{Row: 21, Col: 0, Content: strings.Repeat("-", 80)},                   // ASCII only
		go3270.Field{Row: 22, Col: 0, Content: "Press Enter to save, PF3 Exit, Enter # of new URL:"},
		go3270.Field{Row: 22, Col: 51, Write: true, Name: "choice", Content: ""},
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
		22, 52,      // cursor coordinates
		conn, // network connection
	)
	if err != nil {
		return nil, nil, err
	}
	fieldValues = resp.Values
	if fieldValues["choice"] != "" {
		ch := fieldValues["choice"]
		var i int
		if _, err := fmt.Sscanf(ch, "%2d", &i); err == nil {
			fmt.Println(i)
		}
		newURL = rssFeeds[i]
	}
	if fieldValues["newURL"] != "" {
		newURL = fieldValues["newURL"]
	}

	switch resp.AID {
	case go3270.AIDEnter:
		// Go to big screen size transaction
		return rssfeed, newURL, nil
	case go3270.AIDPF3:
		// Exit
		return nil, nil, nil
	default:
		// re-run current transaction
		return rssurl, nil, nil
	}
}
