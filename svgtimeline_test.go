package svgtimeline_test

import (
	_ "embed"
	"testing"
	"time"

	svgtimeline "github.com/aorith/svg-timeline"
)

//go:embed tests/test1.svg
var testSVG1 string

//go:embed tests/test2.svg
var testSVG2 string

type testRow struct {
	eras   []svgtimeline.Era
	events []svgtimeline.Event
}

func TestNewTimeline(t *testing.T) {
	rows1 := []testRow{
		{
			eras: []svgtimeline.Era{
				{Class: "ctl-request", Text: "262_req", Duration: 10 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 50, 0, time.UTC)},
			},
		},
		{
			eras: []svgtimeline.Era{
				{Class: "ctl-bereq", Text: "32783_bereq", Duration: 4 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 53, 0, time.UTC)},
			},
		},
		{
			events: []svgtimeline.Event{
				{Class: "ctl-e-long", Text: "Long", Duration: 10 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 51, 0, time.UTC)},
				{Class: "ctl-e-long", Text: "Short", Duration: 3 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 60, 0, time.UTC)},
			},
		},
		{
			events: []svgtimeline.Event{
				{Class: "ctl-e-fetch", Text: "Fetch", Duration: 1 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 51, 0, time.UTC)},
				{Class: "ctl-e-process", Text: "Process", Duration: 2 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 52, 0, time.UTC)},
			},
		},
		{
			events: []svgtimeline.Event{
				{Class: "ctl-e-beresp", Text: "Beresp", Duration: 2 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 55, 0, time.UTC)},
				{Class: "ctl-e-berespbody", Text: "BerespBody", Duration: 3 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 57, 0, time.UTC)},
			},
		},
	}

	rows2 := []testRow{
		{
			eras: []svgtimeline.Era{
				{Class: "ctl-request", Text: "262_req", Duration: 10 * time.Second},
			},
		},
		{
			eras: []svgtimeline.Era{
				{Class: "ctl-bereq", Text: "32783_bereq", Duration: 4 * time.Second},
			},
		},
		{
			events: []svgtimeline.Event{
				{Class: "ctl-e-long", Text: "Long", Duration: 10 * time.Second},
				{Class: "ctl-e-long", Text: "Short", Duration: 3 * time.Second},
			},
		},
		{
			events: []svgtimeline.Event{
				{Class: "ctl-e-fetch", Text: "Fetch", Duration: 1 * time.Second},
				{Class: "ctl-e-process", Text: "Process", Duration: 2 * time.Second},
			},
		},
		{
			events: []svgtimeline.Event{
				{Class: "ctl-e-beresp", Text: "Beresp", Duration: 2 * time.Second},
				{Class: "ctl-e-berespbody", Text: "BerespBody", Duration: 3 * time.Second},
			},
		},
	}

	rows3 := []testRow{
		{
			events: []svgtimeline.Event{
				{Class: "ctl-e-long", Text: "Long", Duration: 10 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 50, 0, time.UTC)},
				{Class: "ctl-e-long", Text: "Short", Duration: 3 * time.Second},
			},
		},
	}

	tests := []struct {
		name string
		rows []testRow
		want string
	}{
		{
			name: "Timeline with Times",
			rows: rows1,
			want: testSVG1,
		},
		{
			name: "Timeline without Times",
			rows: rows2,
			want: testSVG2,
		},
		{
			name: "Timeline with mixed Times",
			rows: rows3,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := svgtimeline.NewTimeline()
			for _, tr := range tt.rows {
				row := tl.AddRow(30, 5)
				for _, era := range tr.eras {
					row.AddEra(era)
				}
				for _, event := range tr.events {
					row.AddEvent(event)
				}
			}
			svg, _ := tl.Generate(svgtimeline.DefaultTimelineConfig())
			if svg != tt.want {
				t.Errorf("NewTimeline() = %v, want %v", svg, tt.want)
			}
		})
	}
}
