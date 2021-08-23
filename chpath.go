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

/*
const (
	// iota begins at 0
	// Also, not a very intuitive syntax, but basically
	// the constants will all be initialized with '1 << (1 * iota)'
	// and iota is increasing with every usage.
	Ox int64 = 1 << (1 * iota)
	Ow
	Or
	Gx
	Gw
	Gr
	Ux
	Uw
	Ur
)
*/

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

func main() {

	initialize()

	if help {
		usage = strings.ReplaceAll(usage, "```", "")
		fmt.Println(usage)
		os.Exit(0)
	}

	if ver {
		fmt.Printf("chpath version:\n0.1.%s\n", Version)
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

func doReach() {
	// reach is the simplest operation, will also exit here

	parsePath()

	j := len(bread)
	for _, val := range bread {
		setTo := val.Mode
		if j == 1 {
			// the last element, i.e. target, always needs r (0x444=292)
			setTo |= 292
		}
		if val.IsDir {
			// x needs to be set, regardless of whether this is path or target (0x111=73)
			setTo |= 73
		}
		check(os.Chmod(val.PathActual, fs.FileMode(setTo)))

		setToStr := strconv.FormatInt(setTo, 8)
		fmt.Println("set to", setToStr, ":   ", val.PathActual)

		j--
	}

	// --reach is not supposed to be used with other flags, exiting
	os.Exit(0)
}

func doMode() {
	var shifts []int
	var pattern, effPerm int
	var bigX bool

	// the values are basically the left shift x3 of the specified 'rwxX' pattern
	if strings.Contains(mode, "a") {
		shifts = []int{0, 1, 2}
	} else {
		if strings.Contains(mode, "u") {
			shifts = []int{2}
		}
		if strings.Contains(mode, "g") {
			shifts = append(shifts, 1)
		}
		if strings.Contains(mode, "o") {
			shifts = append(shifts, 0)
		}
	}
	if len(shifts) < 1 {
		fmt.Println("The --mode value (", mode, ") does not contain any of the 'ugoa' instructions.")
		os.Exit(41)
	}

	if strings.Contains(mode, "r") {
		pattern += 4 // binary ..100
	}
	if strings.Contains(mode, "w") {
		pattern += 2 // binary ...10
	}
	if strings.Contains(mode, "x") {
		pattern += 1 // binary ....1
	} else if strings.Contains(mode, "X") {
		// can't add 1 for both x and X
		pattern += 1
		bigX = true
	}
	if pattern == 0 {
		fmt.Println("The --mode value (", mode, ") does not specify permissions (rwxX).")
		os.Exit(42)
	}

	// effective permissions
	for _, multiplier := range shifts {
		/*	pattern = XXX
			XXX << (0 * 3) = 000 000 XXX
			XXX << (1 * 3) = 000 XXX 000
			XXX << (2 * 3) = XXX 000 000
			        SUM    = XXX XXX XXX
		*/
		effPerm += pattern << (multiplier * 3)
	}

	if strings.Contains(mode, "+") {

		elements := strings.Split(strings.TrimPrefix(target, "/"), "/")

		if len(elements) < 2 {
			fmt.Println("This tool is BY DESIGN not changing top-level directories (such as /bin, /home, /proc). The path provided (", target, ") seems to consist of only one level.")
			os.Exit(32)
		}

		var currentPath string
		var currentCrumb crumb
		for i := range elements {
			// work the path from LEFT to RIGHT
			if i == 0 {
				fmt.Printf("Skipping the first-level directory '/%s'.\n", elements[i])
				currentPath = "/" + elements[i]
				continue
			}
			currentPath += "/" + elements[i]
			currentCrumb = getCrumb(currentPath)

			setTo := currentCrumb.Mode
			if i == (len(elements) - 1) {
				// the last element, i.e. target
				if !currentCrumb.IsDir && bigX {
					// if target is a file and the permission is X (not x)
					effPerm--
				}
			}
			// every other element is a directory by definition, so X=x
			setTo |= int64(effPerm)
			modeFrom := strconv.FormatInt(currentCrumb.Mode, 8)
			modeTo := strconv.FormatInt(setTo, 8)
			if modeFrom == modeTo {
				fmt.Println("skipping, is", modeFrom, "already :   ", currentCrumb.PathActual)
			} else {
				fmt.Println("changing mode", modeFrom, "to", modeTo, ":   ", currentCrumb.PathActual)
				// actually modifying the permission
				check(os.Chmod(currentCrumb.PathActual, fs.FileMode(setTo)))
			}

			// not 100% sure whether this is required
			currentPath = currentCrumb.PathActual

		}

	} else if strings.Contains(mode, "-") {
		// work the path from RIGHT to LEFT
		parsePath()
		size := len(bread)

		for j := 1; j <= size; j++ {
			crumb := bread[size-j]
			setTo := crumb.Mode
			if j == 1 {
				// the last element, i.e. target
				if !crumb.IsDir && bigX {
					// if target is a file and the permission is X (not x)
					effPerm--
				}
			}
			// every other element is a directory by definition, so X=x
			setTo &^= int64(effPerm)
			modeFrom := strconv.FormatInt(crumb.Mode, 8)
			modeTo := strconv.FormatInt(setTo, 8)
			if modeFrom == modeTo {
				fmt.Println("skipping, is", modeFrom, "already :   ", crumb.PathActual)
			} else {
				fmt.Println("changing mode", modeFrom, "to", modeTo, ":   ", crumb.PathActual)
				// actually modifying the permission
				check(os.Chmod(crumb.PathActual, fs.FileMode(setTo)))
			}

		}
	} else {
		fmt.Println("The --mode value (", mode, ") does not specify an action (+/-).")
		os.Exit(43)
	}

}
