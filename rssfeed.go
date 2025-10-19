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
	"time"

	"github.com/racingmars/go3270"
)

func rssfeed(conn net.Conn, devinfo go3270.DevInfo, rssFeedURL any) (
	go3270.Tx, any, error) {

	currentURL := rssFeedURL.(string)

	// Accept Enter; PF3/Clear exit and PF4 new url.
	pfkeys := []go3270.AID{go3270.AIDEnter, go3270.AIDPF4}
	exitkeys := []go3270.AID{go3270.AIDPF3, go3270.AIDClear}

	headlines, err := fetchHeadlines(currentURL, maxHeadlines)
	if err != nil {
		headlines = []string{fmt.Sprintf("Error fetching feed: %v", err)}
	}

	// Make a local copy of the screen definition that we can append lines to.
	screen := make(go3270.Screen, len(layout))
	copy(screen, layout)

	now := time.Now().UTC().Format("15:04 UTC")
	title := "RSS Feed"
	header := padCenter(title, 80)
	sub := padCenter(currentURL+" - Updated: "+now, 80)

	screen = append(screen,
		go3270.Field{Row: 0, Col: 0, Content: header, Intense: true},
		go3270.Field{Row: 1, Col: 0, Content: sub},
		go3270.Field{Row: 2, Col: 0, Content: strings.Repeat("-", 80)}, // ASCII only
	)

	row := 3
	for i, h := range headlines {
		for _, line := range wrap80(fmt.Sprintf("%2d. %s", i+1, strings.TrimSpace(h)), 80) {
			if row >= 22 { // leave space for footer/input
				break
			}
			screen = append(screen, go3270.Field{Row: row, Col: 0, Content: line})
			row++
		}
		if row >= 22 {
			break
		}
	}

	screen = append(screen,
		go3270.Field{Row: 22, Col: 0, Content: strings.Repeat("-", 80)}, // ASCII only
		go3270.Field{Row: 23, Col: 0, Content: "Enter=refresh, PF3/Clear=exit, PF4=Change RSS URL"},
	)

	resp, err := go3270.HandleScreenAlt(
		screen,     // the screen to display
		nil,        // (no) rules to enforce
		nil,        // pre-populated values in fields
		pfkeys,     // keys we accept -- validating
		exitkeys,   // keys we accept -- non-validating
		"errormsg", // name of field to put error messages in
		0, 0,       // cursor coordinates
		conn,    // network connection
		devinfo, // device info for alternate screen size support
	)
	if err != nil {
		return nil, nil, err
	}

	switch resp.AID {
	case go3270.AIDEnter:
		// Re-run current transaction, echoing back input
		return rssfeed, currentURL, err
	case go3270.AIDPF4:
		// Go to default screen size transaction
		return rssurl, nil, nil
	case go3270.AIDPF3:
		// Exit
		return nil, nil, nil
	default:
		// re-run current transaction
		return rssfeed, nil, nil
	}
}
