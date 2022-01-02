package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strings"
)

//go:embed README.md
var usage string

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var owner, group, mode, target, Version string
var ver, help, reach bool

func main() {

	initialize()

	if help {
		usage = strings.ReplaceAll(usage, "```", "")
		fmt.Println(usage)
		os.Exit(0)
	}

	if ver {
		fmt.Printf("chpath version:\n0.2.%s\n", Version)
		os.Exit(0)
	}

	if reach {
		doReach()
		// this one exits at the end
	}

	if mode != "" {
		doMode()
	}

}

func doMode() {

	if strings.Contains(mode, "+") {
		addMode()
	} else if strings.Contains(mode, "-") {
		delMode()
	} else {
		fmt.Println("The --mode value (", mode, ") does not specify an action (+/-).")
		os.Exit(43)
	}

}

func doReach() {
	fmt.Println("Setting permissions to reach\n    ", target)
	mode = "a+XR"
	doMode()
	// --reach is not supposed to be used with other flags, exiting
	os.Exit(0)
}
