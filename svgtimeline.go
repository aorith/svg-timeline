// SPDX-License-Identifier: MIT

package svgtimeline

import (
	"fmt"
	"math"
	"strings"
	"time"

	_ "embed"
)

//go:embed default.css
var DefaultStyle string

//go:embed defs.xml
var defs string

type EventType int

const (
	EventTypeTask EventType = iota // A discrete unit of work rendered as a rectangle within its row
	EventTypeEra                   // A time period that spans vertically across all rows below it
)

// Event represents a timeline event
type Event struct {
	Type     EventType     // type of the event - affects how it is drawn on the timeline
	ID       string        // unique HTML identifier
	Class    string        // CSS class name
	Text     string        // text displayed inside of the event rectangle if the event duration provides sufficient width
	Title    string        // tooltip text
	Duration time.Duration // event duration
	Time     time.Time     // absolute start time (leave zero for auto positioning by last duration)
}

// Row represents a row in the timeline
type Row struct {
	height          int
	separatorHeight int
	events          []Event
}

// Timeline represents the entire timeline
type Timeline struct {
	rows []*Row
}

// TimelineConfig represents a timeline config
type TimelineConfig struct {
	ID       string // ID of the SVG
	Width    int
	NumTicks int
	Margins  *[4]int // Margins: top, right, bottom, left
	Style    string  // CSS styling (for reference use DefaultStyle)
}

// DefaultTimelineConfig returns the default TimelineConfig
func DefaultTimelineConfig() TimelineConfig {
	return TimelineConfig{
		ID:       "",
		Width:    1000,
		NumTicks: 8,
		Margins:  &[4]int{15, 30, 15, 10},
		Style:    DefaultStyle,
	}
}

// NewTimeline creates a new timeline
func NewTimeline() *Timeline {
	return &Timeline{
		rows: make([]*Row, 0),
	}
}

// AddRow adds a new row to the timeline
func (t *Timeline) AddRow(height int, separatorHeight int) *Row {
	row := &Row{
		height:          height,
		separatorHeight: separatorHeight,
		events:          make([]Event, 0),
	}
	t.rows = append(t.rows, row)
	return row
}

// GetRows returns the timeline rows
func (t *Timeline) GetRows() []*Row {
	return t.rows
}

// GetLastRow returns the last row
func (t *Timeline) GetLastRow() *Row {
	if len(t.rows) == 0 {
		return nil
	}
	return t.rows[len(t.rows)-1]
}

// MaxDuration returns the maximum duration across all rows
func (t *Timeline) MaxDuration() time.Duration {
	var m time.Duration
	for _, row := range t.rows {
		duration := row.TotalDuration(t.StartTime())
		if duration > m {
			m = duration
		}
	}
	return m
}

// TotalRowHeight calculates the total height of all rows including separators
func (t *Timeline) TotalRowHeight() int {
	total := 0
	for _, row := range t.rows {
		total += row.height + row.separatorHeight
	}
	return total
}

// StartTime returns the earliest time that is currently set on the timeline
// given the existing rows and events
func (t *Timeline) StartTime() time.Time {
	var earliest time.Time
	for _, r := range t.rows {
		rowStartTime := r.StartTime()
		if earliest.IsZero() || rowStartTime.Before(earliest) {
			earliest = rowStartTime
		}
	}
	return earliest
}

// EndTime returns the latest time that is currently set on the timeline
// given the added rows and events (including their durations)
func (t *Timeline) EndTime() time.Time {
	var end time.Time
	for _, r := range t.rows {
		rowEndTime := r.EndTime()
		if end.IsZero() || rowEndTime.After(end) {
			end = rowEndTime
		}
	}
	return end
}

// Generate creates the timeline with the given config
func (t *Timeline) Generate(cfg TimelineConfig) (string, error) {
	earliest, err := checkEvents(t)
	if err != nil {
		return "", err
	}

	const tickHeight = 5
	const tickLabelMargin = 15

	defaults := DefaultTimelineConfig()

	if cfg.Width == 0 {
		cfg.Width = defaults.Width
	}
	if cfg.NumTicks == 0 {
		cfg.NumTicks = defaults.NumTicks
	}
	if cfg.Margins == nil {
		cfg.Margins = defaults.Margins
	}

	marginTop := cfg.Margins[0]
	marginRight := cfg.Margins[1]
	marginBottom := cfg.Margins[2]
	marginLeft := cfg.Margins[3]

	maxDuration := t.MaxDuration()
	totalHeight := t.TotalRowHeight()
	svgHeight := totalHeight + marginTop + marginBottom + tickHeight + tickLabelMargin

	contentWidth := cfg.Width - marginLeft - marginRight
	timelineY := marginTop + totalHeight + tickHeight

	var sb strings.Builder

	// SVG header
	sb.WriteString("<svg")
	if cfg.ID != "" {
		sb.WriteString(fmt.Sprintf(` id="%s"`, escapeXML(cfg.ID)))
	}
	sb.WriteString(fmt.Sprintf(
		` xmlns="http://www.w3.org/2000/svg" width="%[1]d" height="%[2]d" viewBox="0 0 %[1]d %[2]d">`,
		cfg.Width, svgHeight,
	))
	sb.WriteString("\n")

	sb.WriteString("<defs>\n" + defs + "\n")
	if cfg.Style != "" {
		// Optional style
		sb.WriteString("<style>\n" + cfg.Style + "</style>\n")
	}
	sb.WriteString("</defs>\n")

	// Background
	sb.WriteString(fmt.Sprintf(`<rect class="tl-bg" x="0" y="0" width="%d" height="%d" fill="none" />`,
		cfg.Width, svgHeight))

	// Draw rows
	currentY := marginTop
	for _, row := range t.rows {
		if maxDuration <= 0 {
			break
		}
		var currentDuration time.Duration

		// Draw events
		for _, event := range row.events {
			switch event.Type {
			case EventTypeTask:
				if !earliest.IsZero() {
					currentDuration = event.Time.Sub(earliest)
				}

				startX := float64(marginLeft) + float64(contentWidth)*float64(currentDuration)/float64(maxDuration)
				eventWidth := float64(contentWidth) * float64(event.Duration) / float64(maxDuration)

				sb.WriteString("<g")
				if event.ID != "" {
					sb.WriteString(fmt.Sprintf(` id="%s"`, event.ID))
				}
				if event.Class == "" {
					sb.WriteString(` class="tl-event"`)
				} else {
					sb.WriteString(fmt.Sprintf(` class="tl-event %s"`, event.Class))
				}
				sb.WriteString(">\n")

				if event.Title != "" {
					sb.WriteString(fmt.Sprintf(`<title>%s</title>`, escapeXML(event.Title)))
				}

				sb.WriteString(fmt.Sprintf(`<rect x="%f" y="%d" width="%f" height="%d" />`,
					startX, currentY, eventWidth, row.height))
				sb.WriteString("\n")

				// Draw event text
				if event.Text != "" && eventWidth > float64(len(event.Text)*5) { // pixels per char
					textSize := int(max(8, min(float64(row.height/3)+2, float64(eventWidth/4)+3)))
					textX := startX + eventWidth/2
					textY := float64(currentY) + (float64(row.height / 2))

					sb.WriteString(fmt.Sprintf(`<text x="%f" y="%f" font-family="monospace" font-size="%d" dominant-baseline="middle" text-anchor="middle">%s</text>`,
						textX, textY, textSize, escapeXML(event.Text)))
					sb.WriteString("\n")
				}

				sb.WriteString("</g>\n")

				if earliest.IsZero() {
					currentDuration += event.Duration
				}

			case EventTypeEra:
				eraHeight := svgHeight - currentY - marginBottom - (tickHeight * 2) + 2

				if !earliest.IsZero() {
					currentDuration = event.Time.Sub(earliest)
				}

				startX := float64(marginLeft) + float64(contentWidth)*float64(currentDuration)/float64(maxDuration)
				eraWidth := float64(contentWidth) * float64(event.Duration) / float64(maxDuration)

				sb.WriteString("<g")
				if event.ID != "" {
					sb.WriteString(fmt.Sprintf(` id="%s"`, event.ID))
				}
				if event.Class == "" {
					sb.WriteString(` class="tl-era"`)
				} else {
					sb.WriteString(fmt.Sprintf(` class="tl-era %s"`, event.Class))
				}
				sb.WriteString(">\n")

				// NOTE: using a 'hack' to set only the left & right borders: stroke-dasharray="0, <width>, <height>, 0"
				sb.WriteString(fmt.Sprintf(`<rect x="%f" y="%d" width="%f" height="%d" stroke-dasharray="0,%[3]f,%[4]d,0" />`,
					startX, currentY, eraWidth, eraHeight))
				sb.WriteString("\n")

				// Draw era text
				if event.Text != "" && eraWidth > float64(len(event.Text)*5) { // pixels per char
					textSize := int(max(8, min(float64(row.height/3)+2, float64(eraWidth/4)+3)))
					textX := startX + eraWidth/2
					textY := float64(currentY) + float64(row.height)/3

					sb.WriteString(fmt.Sprintf(`<text x="%f" y="%f" font-family="monospace" font-size="%d" dominant-baseline="middle" text-anchor="middle">%s</text>`,
						textX, textY, textSize, escapeXML(event.Text)))
					sb.WriteString("\n")
				}

				sb.WriteString("</g>\n")

				if earliest.IsZero() {
					currentDuration += event.Duration
				}
			}
		}

		currentY += row.height + row.separatorHeight
	}

	// Draw timeline axis
	sb.WriteString(fmt.Sprintf(`<line class="tl-axis" x1="%d" y1="%d" x2="%d" y2="%d"/>`,
		marginLeft, timelineY, marginLeft+contentWidth, timelineY))
	sb.WriteString("\n")

	// Draw tick marks and labels
	sb.WriteString(`<g class="tl-ticks">`)
	sb.WriteString("\n")
	if cfg.NumTicks > 0 && maxDuration > 0 {
		tickDuration := maxDuration / time.Duration(cfg.NumTicks)

		for i := 0; i <= cfg.NumTicks; i++ {
			currentDuration := tickDuration * time.Duration(i)
			x := float64(marginLeft) + float64(contentWidth)*float64(currentDuration)/float64(maxDuration)

			// Tick mark
			sb.WriteString(fmt.Sprintf(`<line x1="%f" y1="%d" x2="%f" y2="%d"/>`,
				x, timelineY-tickHeight, x, timelineY+tickHeight))
			sb.WriteString("\n")

			// Tick label
			label := formatDuration(currentDuration, 2)
			sb.WriteString(fmt.Sprintf(`<text x="%f" y="%d" font-family="monospace" font-size="12" text-anchor="middle">%s</text>`,
				x, timelineY+tickHeight+tickLabelMargin, label))
			sb.WriteString("\n")
		}
	}
	sb.WriteString("</g>\n")

	sb.WriteString("</svg>")
	return sb.String(), nil
}

// AddEvent adds an event to a row
func (r *Row) AddEvent(e Event) {
	r.events = append(r.events, e)
}

// TotalDuration returns the total duration for a row
func (r *Row) TotalDuration(earliest time.Time) time.Duration {
	var total time.Duration
	var maxByTime time.Duration

	for _, event := range r.events {
		total += event.Duration
		if !earliest.IsZero() && !event.Time.IsZero() {
			byTime := event.Time.Sub(earliest) + event.Duration
			if byTime > maxByTime {
				maxByTime = byTime
			}
		}
	}
	return max(total, maxByTime)
}

// StartTime returns the earliest time that is currently set on the row
// given the existing events
func (r *Row) StartTime() time.Time {
	var earliest time.Time
	for _, e := range r.events {
		if e.Time.IsZero() {
			continue
		}
		if earliest.IsZero() || e.Time.Before(earliest) {
			earliest = e.Time
		}
	}
	return earliest
}

// EndTime returns the latest time that is currently set on the row
// given the existing events (including their durations)
func (r *Row) EndTime() time.Time {
	var end time.Time
	for _, e := range r.events {
		if e.Time.IsZero() {
			continue
		}
		eventEnd := e.Time.Add(e.Duration)
		if end.IsZero() || eventEnd.After(end) {
			end = eventEnd
		}
	}
	return end
}

// checkEvents is a helper to ensure consistency across events
// - if any event sets its Time, all events must set it and the earliest time is returned
// - at least one event must have a duration greater than 0
func checkEvents(t *Timeline) (time.Time, error) {
	var hasTime, hasNoTime bool
	var earliest time.Time
	var duration time.Duration

	for _, r := range t.rows {
		for _, e := range r.events {
			duration += e.Duration
			if e.Time.IsZero() {
				hasNoTime = true
			} else {
				hasTime = true
			}
		}
	}

	if hasTime && hasNoTime {
		return earliest, fmt.Errorf("if 'Time' is set on any Event, it must be set on all of them")
	}

	if duration == 0 {
		return earliest, fmt.Errorf("none of the events has a positive duration")
	}

	earliest = t.StartTime()
	return earliest, nil
}

// escapeXML escapes special XML characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// formatDuration rounds a time.Duration to the given digits and returns its String()
func formatDuration(d time.Duration, digits int) string {
	div := time.Duration(math.Pow(10, float64(digits)))
	switch {
	case d > time.Second:
		d = d.Round(time.Second / div)
	case d > time.Millisecond:
		d = d.Round(time.Millisecond / div)
	case d > time.Microsecond:
		d = d.Round(time.Microsecond / div)
	case d > time.Nanosecond:
		d = d.Round(time.Nanosecond / div)
	}
	return d.String()
}
