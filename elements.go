// SPDX-License-Identifier: MIT

package svgtimeline

import (
	"encoding/xml"
)

type svg struct {
	XMLName             xml.Name `xml:"svg"`
	ID                  string   `xml:"id,attr,omitempty"`
	Class               string   `xml:"class,attr,omitempty"`
	Xmlns               string   `xml:"xmlns,attr"`
	Width               string   `xml:"width,attr"`
	Height              string   `xml:"height,attr"`
	ViewBox             string   `xml:"viewBox,attr"`
	PreserveAspectRatio string   `xml:"preserveAspectRatio,attr"`
	Elements            []any    `xml:",any"`
}

type svgDefs struct {
	XMLName  xml.Name `xml:"defs"`
	Elements []any    `xml:",any"`
}

type svgStyle struct {
	XMLName xml.Name `xml:"style"`
	Content string   `xml:",chardata"`
}

type g struct {
	XMLName  xml.Name `xml:"g"`
	ID       string   `xml:"id,attr,omitempty"`
	Class    string   `xml:"class,attr,omitempty"`
	Elements []any    `xml:",any"`
}

type rect struct {
	XMLName         xml.Name `xml:"rect"`
	ID              string   `xml:"id,attr,omitempty"`
	Class           string   `xml:"class,attr,omitempty"`
	X               float64  `xml:"x,attr"`
	Y               float64  `xml:"y,attr"`
	Width           float64  `xml:"width,attr"`
	Height          float64  `xml:"height,attr"`
	Fill            string   `xml:"fill,attr,omitempty"`
	FillOpacity     float64  `xml:"fill-opacity,attr,omitempty"`
	Stroke          string   `xml:"stroke,attr,omitempty"`
	StrokeWidth     int      `xml:"stroke-width,attr,omitempty"`
	StrokeDasharray string   `xml:"stroke-dasharray,attr,omitempty"`
}

type line struct {
	XMLName         xml.Name `xml:"line"`
	ID              string   `xml:"id,attr,omitempty"`
	Class           string   `xml:"class,attr,omitempty"`
	X1              float64  `xml:"x1,attr"`
	Y1              float64  `xml:"y1,attr"`
	X2              float64  `xml:"x2,attr"`
	Y2              float64  `xml:"y2,attr"`
	Fill            string   `xml:"fill,attr,omitempty"`
	Stroke          string   `xml:"stroke,attr,omitempty"`
	StrokeWidth     int      `xml:"stroke-width,attr,omitempty"`
	StrokeDasharray string   `xml:"stroke-dasharray,attr,omitempty"`
	MarkerStart     string   `xml:"marker-start,attr,omitempty"`
	MarkerEnd       string   `xml:"marker-end,attr,omitempty"`
}

type text struct {
	XMLName          xml.Name `xml:"text"`
	ID               string   `xml:"id,attr,omitempty"`
	Class            string   `xml:"class,attr,omitempty"`
	X                float64  `xml:"x,attr"`
	Y                float64  `xml:"y,attr"`
	Fill             string   `xml:"fill,attr,omitempty"`
	Stroke           string   `xml:"stroke,attr,omitempty"`
	FontSize         string   `xml:"font-size,attr,omitempty"`
	FontFamily       string   `xml:"font-family,attr,omitempty"`
	TextAnchor       string   `xml:"text-anchor,attr,omitempty"`
	DominantBaseline string   `xml:"dominant-baseline,attr,omitempty"`
	WritingMode      string   `xml:"writing-mode,attr,omitempty"`
	Transform        string   `xml:"transform,attr,omitempty"`
	Content          string   `xml:",chardata"`
}

type title struct {
	XMLName xml.Name `xml:"title"`
	Content string   `xml:",chardata"`
}
