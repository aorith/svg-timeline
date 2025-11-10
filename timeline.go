// SPDX-License-Identifier: MIT

package svgtimeline

import (
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	_ "embed"
)

//go:embed default.css
var DefaultStyle string

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

	id           string
	width        string
	height       string
	precision    float64
	numTicks     int
	tickHeight   int
	marginTop    int
	marginBottom int
	marginLeft   float64
	marginRight  float64
	style        string

	earliest        time.Time // Earliest time within the timeline
	maxDuration     time.Duration
	tickLabelMargin int
	contentHeight   int
	totalHeight     int
	contentWidth    float64
	totalWidth      float64
}

// NewTimeline creates a new timeline with default config
func NewTimeline() *Timeline {
	return &Timeline{
		rows:         make([]*Row, 0),
		id:           "",
		width:        "100%",
		precision:    float64(1000),
		numTicks:     8,
		tickHeight:   5,
		marginTop:    15,
		marginBottom: 15,
		marginLeft:   10,
		marginRight:  30,
		style:        DefaultStyle,
	}
}

// SetID sets the unique HTML identifier of the timeline SVG
func (t *Timeline) SetID(id string) {
	t.id = id
}

// SetPrecision sets the precision of the timeline
//
// Higher precision creates wider timelines (default: 1000).
func (t *Timeline) SetPrecision(p int) {
	t.precision = float64(p)
}

// SetWidth sets the SVG width.
//
// Any CSS value for size is valid, including pixels or percentages.
func (t *Timeline) SetWidth(width string) {
	t.width = width
}

// SetHeight sets the SVG height.
//
// Any CSS value for size is valid, including pixels or percentages.
func (t *Timeline) SetHeight(height string) {
	t.height = height
}

// SetNumTicks sets the number of ticks for the timeline
func (t *Timeline) SetNumTicks(n int) {
	t.numTicks = n
}

// SetTickHeight sets the height of the timeline ticks
func (t *Timeline) SetTickHeight(h int) {
	t.tickHeight = h
}

// SetMargins sets the margins of the timeline inside of the SVG
func (t *Timeline) SetMargins(top, right, bottom, left int) {
	t.marginTop = top
	t.marginBottom = bottom
	t.marginLeft = float64(left)
	t.marginRight = float64(right)
}

// SetStyle sets the CSS style for the timeline (for reference use the value of DefaultStyle)
func (t *Timeline) SetStyle(s string) {
	t.style = s
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

// GetRowByIndex returns the row at the index or nil if not found
func (t *Timeline) GetRowByIndex(i int) *Row {
	if i >= len(t.rows) {
		return nil
	}
	return t.rows[i]
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

// Generate generates the timeline SVG with the current configuration
func (t *Timeline) Generate() (string, error) {
	err := t.setup()
	if err != nil {
		return "", err
	}

	root := svg{
		Xmlns:               "http://www.w3.org/2000/svg",
		ID:                  t.id,
		Width:               t.width,
		Height:              t.height,
		ViewBox:             fmt.Sprintf("0 0 %f %f", t.totalWidth, float64(t.totalHeight)),
		PreserveAspectRatio: "xMinYMin meet",
	}

	// Definitions
	defs := svgDefs{}
	if t.style != "" {
		defs.Elements = append(defs.Elements, svgStyle{Content: t.style})
	}
	root.Elements = append(root.Elements, defs)

	// Background
	root.Elements = append(root.Elements,
		rect{Class: "tl-bg", X: 0, Y: 0, Width: t.totalWidth, Height: float64(t.totalHeight), Fill: "none"},
	)

	// Draw rows
	currentY := t.marginTop
	for _, row := range t.rows {
		if t.maxDuration <= 0 {
			break
		}
		var currentDuration time.Duration

		// Draw events
		for _, event := range row.events {
			currentDuration = t.drawEvent(&root, event, currentY, row.height, currentDuration)
		}

		currentY += row.height + row.separatorHeight
	}

	// Draw timeline axis
	timelineY := t.marginTop + t.contentHeight + t.tickHeight
	root.Elements = append(root.Elements,
		line{Class: "tl-axis", X1: t.marginLeft, Y1: float64(timelineY), X2: t.marginLeft + t.contentWidth, Y2: float64(timelineY)},
	)

	// Draw tick marks and labels
	group := g{Class: "tl-ticks"}
	if t.numTicks > 0 && t.maxDuration > 0 {
		tickDuration := t.maxDuration / time.Duration(t.numTicks)

		for i := 0; i <= t.numTicks; i++ {
			currentDuration := tickDuration * time.Duration(i)
			x := float64(t.marginLeft) + float64(t.contentWidth)*float64(currentDuration)/float64(t.maxDuration)

			// Tick mark
			topY := timelineY - t.tickHeight
			if i == 0 || i == t.numTicks {
				topY = t.marginTop
			}
			group.Elements = append(group.Elements,
				line{X1: x, Y1: float64(topY), X2: x, Y2: float64(timelineY + t.tickHeight)},
			)

			// Tick label
			label := formatDuration(currentDuration, 2)
			group.Elements = append(group.Elements,
				text{X: x, Y: float64(timelineY + t.tickHeight + t.tickLabelMargin), FontSize: "12", FontFamily: "monospace", TextAnchor: "middle", Content: label},
			)
		}
	}
	root.Elements = append(root.Elements, group)

	var sb strings.Builder
	encoder := xml.NewEncoder(&sb)
	encoder.Indent("", "  ")
	if err := encoder.Encode(root); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// setup initializes timeline variables and ensures consistency across events
// - if any event sets its Time, all events must set it and the earliest time is returned
// - at least one event must have a duration greater than 0
func (t *Timeline) setup() error {
	var hasTime, hasNoTime bool
	var duration time.Duration

	for _, r := range t.rows {
		for _, e := range r.events {
			if e.Duration < 0 {
				return fmt.Errorf("duration of events cannot be negative")
			}
			duration += e.Duration
			if e.Time.IsZero() {
				hasNoTime = true
			} else {
				hasTime = true
			}
		}
	}

	if hasTime && hasNoTime {
		return fmt.Errorf(`when "Time" is set on any Event, it must be set on all of them`)
	}

	if duration == 0 {
		return fmt.Errorf("none of the events has a positive duration")
	}

	// Initialize variables
	t.tickLabelMargin = 15
	t.maxDuration = t.MaxDuration()
	t.contentHeight = t.TotalRowHeight()
	t.earliest = t.StartTime()
	t.totalHeight = t.contentHeight + t.marginTop + t.marginBottom + t.tickHeight + t.tickLabelMargin
	if t.height == "" {
		t.height = strconv.Itoa(t.totalHeight)
	}

	t.contentWidth = min(t.precision, float64(t.maxDuration))
	t.totalWidth = t.contentWidth + t.marginLeft + t.marginRight

	return nil
}

// drawEvent draws an event in the timeline
func (t *Timeline) drawEvent(root *svg, event Event, currentY, rowHeight int, currentDuration time.Duration) time.Duration {
	if !t.earliest.IsZero() {
		currentDuration = event.Time.Sub(t.earliest)
	}

	startX := t.marginLeft + t.contentWidth*float64(currentDuration)/float64(t.maxDuration)
	eventWidth := t.contentWidth * float64(event.Duration) / float64(t.maxDuration)

	var height int
	var strokeDashArray string
	var textYOffset float64

	if event.Type == EventTypeEra {
		height = t.totalHeight - currentY - t.marginBottom - (t.tickHeight * 3)
		strokeDashArray = fmt.Sprintf(`0,%f,%d,0`, eventWidth, height)
		textYOffset = float64(rowHeight) / 3
	} else {
		height = rowHeight
		textYOffset = float64(rowHeight) / 2
	}

	class := "tl-event"
	if event.Type == EventTypeEra {
		class = "tl-era"
	}
	if event.Class != "" {
		class += " " + event.Class
	}

	group := g{ID: event.ID, Class: class}

	// Title
	if event.Title != "" {
		group.Elements = append(group.Elements,
			title{Content: event.Title},
		)
	}

	// Rectangle
	group.Elements = append(group.Elements,
		rect{X: startX, Y: float64(currentY), Width: eventWidth, Height: float64(height), StrokeDasharray: strokeDashArray},
	)

	// Text
	const textWidthFactor = 0.7
	if event.Text != "" {
		textSize := int(min(
			float64(rowHeight/2),
			eventWidth/(float64(len(event.Text))*textWidthFactor),
		))
		if event.Type == EventTypeEra {
			textSize -= 1
		}
		if textSize >= 3 {
			textX := startX + eventWidth/2
			textY := float64(currentY) + textYOffset

			group.Elements = append(group.Elements,
				text{X: textX, Y: textY, FontSize: strconv.Itoa(textSize), FontFamily: "monospace", DominantBaseline: "middle", TextAnchor: "middle", Content: event.Text},
			)
		}
	}

	root.Elements = append(root.Elements, group)

	if t.earliest.IsZero() {
		currentDuration += event.Duration
	}

	return currentDuration
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
