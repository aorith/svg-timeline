// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"os"

	svgtimeline "github.com/aorith/svg-timeline"
)

func main() {
	var (
		inputFile  = flag.String("i", "", "Input CFG file (required)")
		cssFile    = flag.String("s", "", "CSS style file (optional)")
		outputFile = flag.String("o", "", "Output SVG file (default: stdout)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -i <input.cfg> [-s <style.css>] [-o <output.svg>]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Generate SVG timeline from CFG file.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -i timeline.cfg -s style.css -o timeline.svg\n", os.Args[0])
	}

	flag.Parse()

	if *inputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	svg, err := svgtimeline.GenerateFromCFG(*inputFile, *cssFile)
	if err != nil {
		panic(err)
	}

	// Write output
	if *outputFile == "" {
		fmt.Println(svg)
	} else {
		if err := os.WriteFile(*outputFile, []byte(svg), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Timeline written to %s\n", *outputFile)
	}
}
