package main

import (
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

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
		resultBigX = true
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
	var fnAddMode fnCrumb

	appendMask, bigX, bigR = calculateMode()

	fnAddMode = func(thisCrumb crumb) {
		effectiveMode := thisCrumb.Mode
		if thisCrumb.IsTarget {
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
	}

	pathL2R(fnAddMode)

}

func delMode() {
	var negativeMask int
	var bigX bool
	var fnDelMode fnCrumb

	negativeMask, bigX, _ = calculateMode()

	fnDelMode = func(thisCrumb crumb) {
		effectiveMode := thisCrumb.Mode
		// unlike addMode(), we can't change the mask variable directly, copying
		subtractMask := negativeMask
		if thisCrumb.IsTarget {
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

	pathR2L(fnDelMode)

}
