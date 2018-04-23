package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// TODO: Look at using cobra cmd for this.
	flag.Usage = usage
	verbose := flag.Bool("verbose", false, "verbose logging mode")
	enum := flag.Bool("enum-to-check", false, "convert enums to check constraints")
	flag.Parse()

	if len(flag.Args()) < 2 {
		usage()
		os.Exit(1)
	}

	src := flag.Args()[0]
  dst := flag.Args()[1]

  if len(src) == 0 || len(dst) == 0 {
		usage()
		os.Exit(1)
	}

	rr := newRadRoach(src, dst, verboseLogging(*verbose), enumToCheck(*enum))
	rr.run()
}

func usage() {
	fmt.Println("Usage: radroach [FLAGS...] SOURCE_MYSQL_DUMP DESTINATION_CRDB_DUMP")
	flag.PrintDefaults()
}
