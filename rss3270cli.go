// This file is part of https://github.com/MortenHarding/rss3270cli/
// Copyright 2025 by Morten Harding, licensed under the MIT license. See
// LICENSE in the project root for license information.

// It is based on example5 of https://github.com/racingmars/go3270/
// Copyright 2025 by Matthew R. Wilson
// and the code in https://github.com/ErnieTech101/rss3270svr
// Copyright ErnieTech101

package main

import (
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	go3270 "github.com/racingmars/go3270"
	"github.com/subosito/shorturl"
)

type rss struct {
	Channel struct {
		Title string    `xml:"title"`
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}
type rssItem struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

const (
	httpTimeout  = 10 * time.Second
	maxHeadlines = 18 // fits 24x80 with header/footer
)

var layout = go3270.Screen{}
var rssFeeds = readRssUrlFile("rssfeed.url")
var defrssFeedURL = rssFeeds[0]
var rssChannels [20][2]string

func main() {

	//rssChannels := make([][]string,0)

	for i, url := range rssFeeds {
		ct := fetchTitle(url)
		if ct != "" {
			rssChannels[i][0] = ct
		} else {
			rssChannels[i][0] = "No Title found"
		}
		rssChannels[i][1] = url
		i++
	}

	//Define command line arguments
	port := flag.String("port", "7300", "Listen on port")
	flag.Parse()
	listenAddr := ":" + *port

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic(err)
	}
	fmt.Println("LISTENING ON PORT " + listenAddr + " FOR CONNECTIONS")
	fmt.Println("Press Ctrl-C to end server.")
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		go handle(conn)
	}
}

// handle is the handler for individual user connections.
func handle(conn net.Conn) {
	defer conn.Close()

	//Log the client IP connection
	connectTime := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	clientAddress := conn.RemoteAddr().String()
	fmt.Println(connectTime + " - connection from " + clientAddress)
	// Always begin new connection by negotiating the telnet options
	devinfo, err := go3270.NegotiateTelnet(conn)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = go3270.RunTransactions(conn, devinfo, rssfeed, defrssFeedURL)
	if err != nil {
		fmt.Println(err)
	}
	disconnectconnectTime := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
	fmt.Println(disconnectconnectTime + " - disconnect from " + clientAddress)
}

func fetchTitle(url string) string {
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var r rss
	if err := xml.NewDecoder(resp.Body).Decode(&r); err != nil {
		fmt.Println(err)
	}
	title := replaceUnhandledChar(r.Channel.Title)

	if title == "" {
		title = "No Title found"
	}
	return title
}

func fetchHeadlines(url string, limit int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var r rss
	if err := xml.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	out := make([]string, 0, limit)
	for _, it := range r.Channel.Items {
		t := strings.TrimSpace(it.Title)
		rpl := replaceUnhandledChar(t)

		if rpl != "" {
			out = append(out, rpl)
			if len(out) >= limit {
				break
			}
		}
	}
	if len(out) == 0 {
		out = []string{"(No headlines found)"}
	}
	return out, nil
}

func fetchHeadlineLinks(url string, limit int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var r rss
	if err := xml.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	out := make([]string, 0, limit)
	for _, it := range r.Channel.Items {
		t := strings.TrimSpace(it.Title)
		rpl := replaceUnhandledChar(t)
		strleng := 45
		var str string

		//add the url link for the item to the output
		l := strings.TrimSpace(it.Link)
		if l != "" {
			provider := "tinyurl"
			u, err := shorturl.Shorten(it.Link, provider)
			if err == nil {
				if len([]rune(t)) > strleng {
					str = rpl[0:strleng] + " " + string(u)
				} else {
					str = padRight(rpl, strleng) + " " + string(u)
				}
			}
		} else {
			str = rpl
		}

		if str != "" {
			out = append(out, str)
			if len(out) >= limit {
				break
			}
		}
	}
	if len(out) == 0 {
		out = []string{"(No headlines found)"}
	}
	return out, nil
}

func wrap80(s string, width int) []string {
	var lines []string
	s = strings.ReplaceAll(s, "\n", " ")
	for len(s) > width {
		cut := width
		if idx := strings.LastIndex(s[:width], " "); idx > 0 {
			cut = idx
		}
		lines = append(lines, padRight(s[:cut], width))
		s = strings.TrimSpace(s[cut:])
	}
	lines = append(lines, padRight(s, width))
	return lines
}

func max80(s string, width int) []string {
	var lines []string
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > width {
		cut := width
		if idx := strings.LastIndex(s[:width], " "); idx > 0 {
			cut = idx
		}
		lines = append(lines, padRight(s[:cut], width))
	} else {
		lines = append(lines, padRight(s, width))
	}

	return lines
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

func padCenter(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	left := (w - len(s)) / 2
	right := w - len(s) - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func replaceUnhandledChar(s string) string {
	//Define characters that must be replaced
	//go3270 will do the conversion to EBCDIC
	//so this is a conversion from
	// utf8 -> ascii
	r := strings.NewReplacer(
		"å", "aa",
		"ø", "oe",
		"æ", "ae",
		"Å", "AA",
		"Ø", "OE",
		"Æ", "AE",
		"–", "-",
		"’", "'",
		"‘", "'",
		"`", "'",
		"»", "'",
		"«", "'",
		"€", "",
		"—", "-",
		"\u00ad", "", //Soft hyphen
		"é", "e",
		"Ö", "Oe",
		"ö", "oe",
		"Ä", "Ae",
		"ä", "ae",
		"Á", "A",
		"ö", "oe",
		"ä", "ae",
		"ü", "ue",
	)

	line := r.Replace(s)

	return line
}

func readRssUrlFile(filename string) []string {
	content, err := os.ReadFile(filename)
	lines := strings.Split(string(content), "\n")
	out := []string{}

	if err != nil {
		fmt.Println(err)
	}
	// Only return lines with 'http'
	for _, line := range lines {
		if strings.Count(line, "http") == 1 {
			out = append(out, line)
		}
	}

	return out

}
