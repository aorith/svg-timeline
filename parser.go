// SPDX-License-Identifier: MIT

package svgtimeline

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// GenerateFromCFG generates the timeline by parsing a config file with an optional css style
func GenerateFromCFG(filename string, cssFilename string) (string, error) {
	var cssStyle string
	if cssFilename != "" {
		css, err := os.ReadFile(cssFilename)
		if err != nil {
			return "", fmt.Errorf("error reading file '%s': %v", cssFilename, err)
		}
		cssStyle = string(css)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("error reading file '%s': %v", filename, err)
	}

	r := bytes.NewReader(data)
	scanner := bufio.NewScanner(r)

	// Initialize the timeline
	tl := NewTimeline()

	margins := [4]int{0, 0, 0, 0} // top , right , bottom , left
	setMargins := false
	var currentEvent *Event

	currentSection := ""
	lineNum := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		// Skip empty lines and comments
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.Split(line, " ")

		switch line[0] {
		case '@':
			if currentEvent != nil {
				row := tl.GetLastRow()
				if row == nil {
					return "", fmt.Errorf("error at line %d, cannot add an event without creating a row first", lineNum)
				}
				row.AddEvent(*currentEvent)
				currentEvent = nil
			}

			currentSection = parts[0] // @timeline, @row, @task, @era
			switch currentSection {
			case "@row":
				height := parseIntDefault(parts, 1, 30)
				separator := parseIntDefault(parts, 2, 5)
				tl.AddRow(height, separator)
			case "@era":
				currentEvent = &Event{Type: EventTypeEra}
			case "@task":
				currentEvent = &Event{Type: EventTypeTask}
			}

		default:
			key, val, ok := strings.Cut(line, "=")
			if ok {
				key = strings.TrimSpace(key)
				val = strings.TrimSpace(val)
			} else {
				return "", fmt.Errorf("unknown value at line %d", lineNum)
			}

			switch currentSection {
			case "@timeline":
				switch key {

				// Single digit properties
				case "width", "num_ticks", "tick_height", "margin_top", "margin_bottom", "margin_left", "margin_right":
					x, err2 := strconv.Atoi(val)
					if err2 != nil {
						return "", fmt.Errorf("error at line %d: %v", lineNum, err2)
					}

					switch key {
					case "width":
						tl.SetWidth(x)
					case "num_ticks":
						tl.SetNumTicks(x)
					case "tick_height":
						tl.SetTickHeight(x)
					case "margin_top":
						setMargins = true
						margins[0] = x
					case "margin_right":
						setMargins = true
						margins[1] = x
					case "margin_bottom":
						setMargins = true
						margins[2] = x
					case "margin_left":
						setMargins = true
						margins[3] = x
					}

				case "id":
					tl.SetID(val)

				default:
					return "", fmt.Errorf("unknown property '%s' at line %d", key, lineNum)
				}

			case "@row":
				return "", fmt.Errorf("error at line %d, row has no configuration options", lineNum)

			case "@task", "@era":
				switch key {
				case "id":
					currentEvent.ID = val

				case "class":
					currentEvent.Class = val

				case "text":
					currentEvent.Text = val

				case "title":
					currentEvent.Title = val

				case "duration":
					dur, err2 := time.ParseDuration(val)
					if err2 != nil {
						return "", fmt.Errorf("error at line %d while parsing duration of event, %v", lineNum, err2)
					}
					currentEvent.Duration = dur

				case "time":
					t, err2 := parseTime(val)
					if err2 != nil {
						return "", err2
					}
					currentEvent.Time = t

				default:
					return "", fmt.Errorf("unknown event property '%s' at line %d", key, lineNum)
				}

			default:
				return "", fmt.Errorf("unknown section: %s", currentSection)
			}
		}

	}

	if err = scanner.Err(); err != nil {
		return "", fmt.Errorf("scanner error: %v", err)
	}

	// Last event
	row := tl.GetLastRow()
	if row == nil {
		return "", fmt.Errorf("error at line %d, cannot add an event without creating a row first", lineNum)
	}
	row.AddEvent(*currentEvent)
	currentEvent = nil

	if setMargins {
		tl.SetMargins(margins[0], margins[1], margins[2], margins[3])
	}

	if cssStyle != "" {
		tl.SetStyle(cssStyle)
	}

	svg, err := tl.Generate()
	if err != nil {
		return "", err
	}

	return svg, nil
}

// parseIntDefault is a helper function to convert a string to int
// returns the default value if parsing fails
func parseIntDefault(parts []string, i, def int) int {
	if len(parts) <= i {
		return def
	}
	n, err := strconv.Atoi(parts[i])
	if err != nil {
		return def
	}
	return n
}

// parseTime tries to parse time strings in common formats
func parseTime(input string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05.99Z", // UTC with nanosecond precision
		time.UnixDate,             // Mon Jan _2 15:04:05 MST 2006
		time.ANSIC,                // Mon Jan _2 15:04:05 2006
		time.RFC3339,              // 2006-01-02T15:04:05Z07:00
		time.RFC1123,              // Mon, 02 Jan 2006 15:04:05 MST
		time.RFC822,               // 02 Jan 06 15:04 MST
		time.RFC850,               // Monday, 02-Jan-06 15:04:05 MST
		time.DateTime,             // 2006-01-02 15:04:05
		"2006/01/02 15:04:05",     // Common slash style
		"02/01/2006 15:04:05",     // European style
		time.DateOnly,             // 2006-01-02
		"02/01/2006",              // DD/MM/YYYY
		"02 Jan 2006",             // Human style
		"02-Jan-2006",             // Human with dashes
		"15:04:05.99",             // With nanosecond precision
		"15:04:05",                // Only time
		"15:04",                   // Hour and minute only
	}

	var t time.Time
	var err error
	for _, layout := range formats {
		t, err = time.Parse(layout, input)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time format: %s\nyou might use one of the following formats: %v", input, formats)
}
