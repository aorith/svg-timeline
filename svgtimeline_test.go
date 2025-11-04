package svgtimeline_test

import (
	_ "embed"
	"fmt"
	"os"
	"testing"
	"time"

	svgtimeline "github.com/aorith/svg-timeline"
)

//go:embed tests/test1.svg
var testSVG1 string

//go:embed tests/test2.svg
var testSVG2 string

type testRow struct {
	events []svgtimeline.Event
}

func TestNewTimeline(t *testing.T) {
	rows1 := []testRow{
		{
			events: []svgtimeline.Event{
				{Type: svgtimeline.EventTypeEra, Class: "ctl-request", Text: "262_req", Duration: 10 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 50, 0, time.UTC)},
			},
		},
		{
			events: []svgtimeline.Event{
				{Type: svgtimeline.EventTypeEra, Class: "ctl-bereq", Text: "32783_bereq", Duration: 4 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 53, 0, time.UTC)},
			},
		},
		{
			events: []svgtimeline.Event{
				{Type: svgtimeline.EventTypeTask, Class: "ctl-e-long", Text: "Long", Duration: 10 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 51, 0, time.UTC)},
				{Type: svgtimeline.EventTypeTask, Class: "ctl-e-long", Text: "Short", Duration: 3 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 60, 0, time.UTC)},
			},
		},
		{
			events: []svgtimeline.Event{
				{Type: svgtimeline.EventTypeTask, Class: "ctl-e-fetch", Text: "Fetch", Duration: 1 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 51, 0, time.UTC)},
				{Type: svgtimeline.EventTypeTask, Class: "ctl-e-process", Text: "Process", Duration: 2 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 52, 0, time.UTC)},
			},
		},
		{
			events: []svgtimeline.Event{
				{Type: svgtimeline.EventTypeTask, Class: "ctl-e-beresp", Text: "Beresp", Duration: 2 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 55, 0, time.UTC)},
				{Type: svgtimeline.EventTypeTask, Class: "ctl-e-berespbody", Text: "BerespBody", Duration: 3 * time.Second, Time: time.Date(2025, 11, 1, 12, 20, 57, 0, time.UTC)},
			},
		},
	}

	rows2 := []testRow{
		{
			events: []svgtimeline.Event{
				{Type: svgtimeline.EventTypeEra, Class: "ctl-request", Text: "262_req", Duration: 10 * time.Second},
			},
		},
		{
			events: []svgtimeline.Event{
				{Type: svgtimeline.EventTypeEra, Class: "ctl-bereq", Text: "32783_bereq", Duration: 4 * time.Second},
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

	rows4 := []testRow{
		{
			events: []svgtimeline.Event{
				{Class: "ctl-e-long", Text: "Long", Duration: 10 * time.Second},
				{Class: "ctl-e-long", Text: "Short", Duration: -3 * time.Second},
			},
		},
	}

	rows5 := []testRow{
		{
			events: []svgtimeline.Event{
				{Class: "ctl-e-long", Text: "Long", Duration: 0},
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
		{
			name: "Timeline with negative duration",
			rows: rows4,
			want: "",
		},
		{
			name: "Timeline with 0 duration",
			rows: rows5,
			want: "",
		},
		{
			name: "Timeline without events",
			rows: []testRow{},
			want: "",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tl := svgtimeline.NewTimeline()
			for _, tr := range tt.rows {
				row := tl.AddRow(30, 5)
				for _, event := range tr.events {
					row.AddEvent(event)
				}
			}
			svg, err := tl.Generate()
			if err != nil {
				fmt.Printf("%v\n", err)
			}
			if svg != tt.want {
				gotFn := fmt.Sprintf("%d_got_test.svg", i)
				wantFn := fmt.Sprintf("%d_want_test.svg", i)
				t.Errorf(`[%s] failed, resulting svg files saved as "%s" and "%s"`, tt.name, gotFn, wantFn)
				_ = os.WriteFile(gotFn, []byte(svg), 0644)
				_ = os.WriteFile(wantFn, []byte(tt.want), 0644)
			}
		})
	}
}
