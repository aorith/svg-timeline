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

// Event represents a timeline event
type Event struct {
	ID       string
	Class    string
	Text     string
	Title    string
	Duration time.Duration
	Offset   time.Duration
}

// Era represents a timeline era
type Era struct {
	ID       string
	Class    string
	Text     string
	Duration time.Duration
	Offset   time.Duration
}

// Row represents a row in the timeline
type Row struct {
	height          int
	separatorHeight int
	events          []Event
	eras            []Era
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

// Timeline represents the entire timeline
type Timeline struct {
	rows []*Row
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

// AddEvent adds an event to a row
func (r *Row) AddEvent(e Event) {
	r.events = append(r.events, e)
}

// AddEra adds an era to a row
func (r *Row) AddEra(e Era) {
	r.eras = append(r.eras, e)
}

// getTotalDuration calculates the total duration for a row
func (r *Row) getTotalDuration() time.Duration {
	var total time.Duration
	for _, era := range r.eras {
		total += era.Duration + era.Offset
	}
	for _, event := range r.events {
		total += event.Duration + event.Offset
	}
	return total
}

// getMaxDuration returns the maximum duration across all rows
func (t *Timeline) getMaxDuration() time.Duration {
	var m time.Duration
	for _, row := range t.rows {
		duration := row.getTotalDuration()
		if duration > m {
			m = duration
		}
	}
	return m
}

// getTotalRowHeight calculates the total height of all rows including separators
func (t *Timeline) getTotalRowHeight() int {
	total := 0
	for _, row := range t.rows {
		total += row.height + row.separatorHeight
	}
	return total
}

// Generate creates the SVG string
func (t *Timeline) Generate(cfg TimelineConfig) string {
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

	maxDuration := t.getMaxDuration()
	totalHeight := t.getTotalRowHeight()
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
		var currentTime time.Duration

		// Draw eras
		for _, era := range row.eras {
			eraHeight := svgHeight - currentY - marginBottom - (tickHeight * 2) + 2
			currentTime += era.Offset

			startX := float64(marginLeft) + float64(contentWidth)*float64(currentTime)/float64(maxDuration)
			eraWidth := float64(contentWidth) * float64(era.Duration) / float64(maxDuration)

			sb.WriteString("<g")
			if era.ID != "" {
				sb.WriteString(fmt.Sprintf(` id="%s"`, era.ID))
			}
			if era.Class == "" {
				sb.WriteString(` class="tl-era"`)
			} else {
				sb.WriteString(fmt.Sprintf(` class="tl-era %s"`, era.Class))
			}
			sb.WriteString(">\n")

			// NOTE: using a 'hack' to set only the left & right borders: stroke-dasharray="0, <width>, <height>, 0"
			sb.WriteString(fmt.Sprintf(`<rect x="%f" y="%d" width="%f" height="%d" stroke-dasharray="0,%[3]f,%[4]d,0" />`,
				startX, currentY, eraWidth, eraHeight))
			sb.WriteString("\n")

			// Draw era text
			if era.Text != "" && eraWidth > float64(len(era.Text)*5) { // pixels per char
				textSize := int(max(8, min(float64(row.height/3)+2, float64(eraWidth/4)+3)))
				textX := startX + eraWidth/2
				textY := float64(currentY) + float64(row.height)/3

				sb.WriteString(fmt.Sprintf(`<text x="%f" y="%f" font-family="monospace" font-size="%d" dominant-baseline="middle" text-anchor="middle">%s</text>`,
					textX, textY, textSize, escapeXML(era.Text)))
				sb.WriteString("\n")
			}

			sb.WriteString("</g>\n")

			currentTime += era.Duration
		}

		// Draw events
		for _, event := range row.events {
			currentTime += event.Offset

			startX := float64(marginLeft) + float64(contentWidth)*float64(currentTime)/float64(maxDuration)
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

			currentTime += event.Duration
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
	return sb.String()
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
