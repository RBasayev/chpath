package main

import (
	"flag"
	"path/filepath"
)

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
