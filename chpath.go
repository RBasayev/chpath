package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed README.md
var usage string

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type crumb struct {
	Level      int
	PathArg    string
	PathActual string
	IsDir      bool
	Mode       int64
}

var owner, group, mode, target, Version string
var ver, help, reach bool
var bread []crumb // populated in parsePath()

func initialize() {
	flag.BoolVar(&ver, "v", false, "show version")
	flag.BoolVar(&ver, "version", false, "")

	flag.BoolVar(&help, "h", false, "display Help")
	flag.BoolVar(&help, "help", false, "")

	flag.StringVar(&owner, "o", "", "NOT IMPLEMENTED YET - change owner, like chown")
	flag.StringVar(&owner, "owner", "", "")

	flag.StringVar(&group, "g", "", "NOT IMPLEMENTED YET - change group, like chgrp")
	flag.StringVar(&group, "group", "", "")

	flag.StringVar(&mode, "m", "", "change mode, like chmod")
	flag.StringVar(&mode, "mode", "", "")

	flag.BoolVar(&reach, "r", false, "change mode along the path to make the target reachable - x along the path, r for the last element")
	flag.BoolVar(&reach, "reach", false, "")

	flag.Parse()

	t := flag.Arg(0)

	target, _ = filepath.Abs(t)
}

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

func getCrumb(dirtyPath string) crumb {
	_, err := os.Stat(dirtyPath)
	if err != nil {
		fmt.Println("The path (", dirtyPath, ") doesn't exist or cannot be reached by the current user.")
		os.Exit(33)
	}

	var b crumb
	var m os.FileInfo
	//var p int64
	b.Level = len(strings.Split(strings.TrimPrefix(dirtyPath, "/"), "/")) - 1
	b.PathArg = dirtyPath
	b.PathActual, _ = filepath.EvalSymlinks(b.PathArg)

	m, err = os.Stat(b.PathActual)
	check(err)
	b.IsDir = m.IsDir()

	if reach || (mode != "") {
		b.Mode = int64(m.Mode().Perm())
	}

	return b
}

func parsePath() {
	_, err := os.Stat(target)
	if err != nil {
		fmt.Println("The target (", target, ") doesn't exist or cannot be reached by the current user.")
		os.Exit(31)
	}

	elements := strings.Split(strings.TrimPrefix(target, "/"), "/")

	if len(elements) < 2 {
		fmt.Println("This tool is BY DESIGN not changing top-level directories (such as /bin, /home, /proc). The path provided (", target, ") seems to consist of only one level.")
		os.Exit(32)
	}

	for i := range elements {
		if i == 0 {
			fmt.Printf("Skipping the first-level directory '/%s'.\n", elements[i])
			continue
		}

		var current string
		current = "/" + strings.Join(elements[0:i+1], "/")

		bread = append(bread, getCrumb(current))
	}

}

func setMode(newMode int64, oldMode int64, filePath string) {
	modeFrom := strconv.FormatInt(oldMode, 8)
	modeTo := strconv.FormatInt(newMode, 8)
	// fmt.Printf("- from: %09s\n", strconv.FormatInt(int64(oldMode), 2))
	// fmt.Printf("-   to: %09s\n", strconv.FormatInt(int64(newMode), 2))
	if modeFrom == modeTo {
		fmt.Println("skipping, is", modeFrom, "already :   ", filePath)
	} else {
		fmt.Println("changing mode", modeFrom, "to", modeTo, ":   ", filePath)
		// actually modifying the permission
		check(os.Chmod(filePath, fs.FileMode(newMode)))
	}

}

func doReach() {
	mode = "a+XR"
	doMode()
	// --reach is not supposed to be used with other flags, exiting
	os.Exit(0)
}

func calculateShifts() []int {
	var result []int
	// the values are basically the left shift x3 of the specified 'rwxX' pattern
	if strings.Contains(mode, "a") {
		result = []int{0, 1, 2}
	} else {
		if strings.Contains(mode, "u") {
			result = []int{2}
		}
		if strings.Contains(mode, "g") {
			result = append(result, 1)
		}
		if strings.Contains(mode, "o") {
			result = append(result, 0)
		}
	}
	if len(result) < 1 {
		fmt.Println("The --mode value (", mode, ") does not contain any of the 'ugoa' instructions.")
		os.Exit(41)
	}

	return result
}

func calculateMode() (int, bool, bool) {
	var pattern, resultMode int
	var resultBigX, resultBigR bool // defaults to 'false'

	if strings.Contains(mode, "r") {
		pattern += 4 // binary ..100
	} else if strings.Contains(mode, "R") {
		// undocumented parameter, used for --reach
		resultBigR = true
	}
	if strings.Contains(mode, "w") {
		pattern += 2 // binary ...10
	}
	if strings.Contains(mode, "x") {
		pattern += 1 // binary ....1
	} else if strings.Contains(mode, "X") {
		// "else if" because we can't add 1 for both x and X
		pattern += 1
	}
	if pattern == 0 {
		fmt.Println("The --mode value (", mode, ") does not specify permissions (rwxX).")
		os.Exit(42)
	}

	// effective permissions
	for _, multiplier := range calculateShifts() {
		/*	pattern = XXX
			XXX << (0 * 3) = 000 000 XXX
			XXX << (1 * 3) = 000 XXX 000
			XXX << (2 * 3) = XXX 000 000
			        SUM    = XXX XXX XXX
		*/
		resultMode += pattern << (multiplier * 3)
	}

	return resultMode, resultBigX, resultBigR
}

func addMode() {
	var appendMask int
	var bigX, bigR bool

	appendMask, bigX, bigR = calculateMode()

	elements := strings.Split(strings.TrimPrefix(target, "/"), "/")

	if len(elements) < 2 {
		fmt.Println("This tool is BY DESIGN not changing top-level directories (such as /bin, /home, /proc). The path provided (", target, ") seems to consist of only one level.")
		os.Exit(32)
	}

	var currentPath string
	var thisCrumb crumb
	for i := range elements {
		// work the path from LEFT to RIGHT
		if i == 0 {
			fmt.Printf("Skipping the first-level directory '/%s'.\n", elements[i])
			currentPath = "/" + elements[i]
			continue
		}
		currentPath += "/" + elements[i]
		thisCrumb = getCrumb(currentPath)

		effectiveMode := thisCrumb.Mode
		if i == (len(elements) - 1) {
			// the last element (LEFT to RIGHT), i.e. target
			if !thisCrumb.IsDir && bigX {
				// if target is a file and the permission is X (not x)
				// we need to remove execute from the mask
				// bit clear (AND NOT) 001 001 001
				appendMask &^= 73
			}
			if bigR {
				// if --reach, we need to make the target readable
				// bit add (bitwise OR)
				appendMask |= 292
			}
		}
		// all other elements are directories by definition, so X=x
		// bit add (bitwise OR)
		effectiveMode |= int64(appendMask)
		setMode(effectiveMode, thisCrumb.Mode, thisCrumb.PathActual)
		// fmt.Printf("- mask: %09s\n", strconv.FormatInt(int64(appendMask), 2))

		// not 100% sure whether this is required
		currentPath = thisCrumb.PathActual

	}

}

func delMode() {
	var negativeMask int
	var bigX bool

	negativeMask, bigX, _ = calculateMode()

	// work the path from RIGHT to LEFT
	parsePath()
	size := len(bread)

	for i := 1; i <= size; i++ {
		thisCrumb := bread[size-i]
		effectiveMode := thisCrumb.Mode
		// unlike addMode(), we can't change the mask variable directly, copying
		subtractMask := negativeMask
		if i == 1 {
			// the first element (RIGHT to LEFT), i.e. target
			if !thisCrumb.IsDir && bigX {
				// if target is a file and the permission is X (not x)
				// we need to remove execute from the mask
				// bit clear (AND NOT) 001 001 001
				subtractMask &^= 73
			}
		}
		// all other elements are directories by definition, so X=x
		// bit clear (AND NOT)
		effectiveMode &^= int64(subtractMask)
		setMode(effectiveMode, thisCrumb.Mode, thisCrumb.PathActual)
		// fmt.Printf("- mask: %09s\n", strconv.FormatInt(int64(subtractMask), 2))

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
