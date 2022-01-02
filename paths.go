package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type crumb struct {
	Level      int
	PathArg    string
	PathActual string
	Mode       int64
	IsDir      bool
	IsTarget   bool
}

type fnCrumb func(crumb)

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

func pathL2R(processCrumb fnCrumb) {

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
		thisCrumb.IsTarget = (i == (len(elements) - 1))

		// this function is implemented in the caller
		// this way I can use pathL2R() with --mode, --owner and --group
		processCrumb(thisCrumb)

		// not 100% sure whether this is required
		currentPath = thisCrumb.PathActual

	}

}

func pathR2L(processCrumb fnCrumb) {
	var bread []crumb

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

		current := "/" + strings.Join(elements[0:i+1], "/")

		bread = append(bread, getCrumb(current))
	}

	size := len(bread)
	for i := 1; i <= size; i++ {
		thisCrumb := bread[size-i]
		if i == 1 {
			thisCrumb.IsTarget = true
		}

		// this function is implemented in the caller
		// it is only used in delMode(), but I figured that it's
		// clearer if I maintain consistency with parseL2R()
		processCrumb(thisCrumb)

	}

}
