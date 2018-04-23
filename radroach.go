package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

var (
	empty = []byte("")
	space = []byte(" ")

	simpleReplacements = map[*regexp.Regexp][]byte{
		// Syntax
		regexp.MustCompile("`"):                           empty,
		regexp.MustCompile(`^.* ENGINE=.*$/\)`):           empty,
		regexp.MustCompile(`\s\sKEY `):                    []byte("  INDEX "),
		regexp.MustCompile(`\s\sFULLTEXT KEY `):           []byte("  INDEX "),
		regexp.MustCompile("ON UPDATE CURRENT_TIMESTAMP"): empty,
		regexp.MustCompile(" ON DELETE CASCADE"):          empty,
		regexp.MustCompile("COMMENT.*,"):                  []byte(","),
		regexp.MustCompile("DEFAULT b'0',$"):              []byte("DEFAULT 0,"),
		regexp.MustCompile("DEFAULT b'1',$"):              []byte("DEFAULT 1,"),
		regexp.MustCompile(`\s\sUNIQUE KEY `):             []byte("  UNIQUE INDEX "),

		// Types
		regexp.MustCompile(" double "):                                  []byte(" FLOAT "),
		regexp.MustCompile(` int\((.*)\) `):                             []byte(" INT "),
		regexp.MustCompile(` double\((.*)\) `):                          []byte(" FLOAT "),
		regexp.MustCompile(` bigint\((.*)\) `):                          []byte(" INT "),
		regexp.MustCompile(` binary\((.*)\) `):                          []byte(" BYTES "),
		regexp.MustCompile(` tinyint\((.*)\) `):                         []byte(" INT "),
		regexp.MustCompile(" datetime "):                                []byte(" TIMESTAMP "),
		regexp.MustCompile(" mediumtext "):                              []byte(" TEXT "),
		regexp.MustCompile(" unsigned "):                                []byte(" "),
		regexp.MustCompile(" mediumtext,"):                              []byte(" TEXT,"),
		regexp.MustCompile(`int DEFAULT '(.*)',$`):                      []byte("INT DEFAULT $1,"),
		regexp.MustCompile(`int NOT NULL DEFAULT '(.*)',$`):             []byte("INT NOT NULL DEFAULT $1,"),
		regexp.MustCompile(`\(decimal(.*)\) NOT NULL DEFAULT '(.*)',$`): []byte("$1 NOT NULL DEFAULT $2,"),
		regexp.MustCompile(`\(decimal(.*)\) DEFAULT '(.*)',$`):          []byte("$1 DEFAULT $2,"),

		// JSON
		regexp.MustCompile(`\\"([A-Za-z0-9\\\-._~:/?#\[\]@!$&'()*+,;=]+)\\"`): []byte(`"$1"`),
		regexp.MustCompile(`\\\\"`):                                           []byte(`"`),
		regexp.MustCompile(`\\"{`):                                            []byte("{"),
		regexp.MustCompile(`}\\"`):                                            []byte("}"),
		regexp.MustCompile(`\\"`):                                             []byte(`"`),
		regexp.MustCompile(`\\'`):                                             []byte(`''`),
	}

	blankLines     = regexp.MustCompile(`\s\s\n`)
	constraints    = regexp.MustCompile("(?smU)CONSTRAINT.*REFERENCES.*$")
	tables         = regexp.MustCompile("(?smU)^(CREATE|INSERT).*;$")
	tableName      = regexp.MustCompile(`CREATE TABLE "(.*)"`)
	trailingCommas = regexp.MustCompile(`,\n\);`)

	enumLine = regexp.MustCompile(`(.*enum.*)`)
	enum     = regexp.MustCompile(` enum\((.*)\) `)
)

// option defines a function which modifies an element of radroach config.
type option func(*radroach)

// radroach defines the config required to roach a mysql dump.
type radroach struct {
	enumToCheck bool
	dst         string
	src         string
	logger      *log.Logger
	verbose     bool
}

// run reads the source file, performs the roach and persists the output.
func (r *radroach) run() {
	// Get info on the source file.
	srcInfo, err := os.Stat(r.src)
	if err != nil {
		r.log(fmt.Errorf("could not stat source file %q: %s", r.src, err))

		fmt.Println("Hmm, couldn't open the source mysql file for reading. Make sure the path and permissions are correct.")
		os.Exit(1)
	}

	// Read the source file.
	input, err := ioutil.ReadFile(r.src)
	if err != nil {
		r.log(fmt.Errorf("could not read file %q: %s", r.src, err))

		fmt.Println("Hmm, couldn't read the source mysql file for reading. Make sure the path and permissions are correct.")
		os.Exit(1)
	}

	// Do the roaching.
	output, err := r.roach(input)
	if err != nil {
		r.log(fmt.Errorf("could not run mysql data: %s", err))

		fmt.Println("Damn, couldn't convert the mysql dump to crdb. Make sure the source file was prepared correctly.")
		os.Exit(1)
	}

	// Persist the output back to disk.
	err = ioutil.WriteFile(r.dst, output, srcInfo.Mode())
	if err != nil {
		r.log(fmt.Errorf("could not save the crdb data to file %q: %s", r.dst, err))

		fmt.Println("Oh dear, couldn't write the crdb data. Make sure you have the correct permissions.")
		os.Exit(1)
	}
}

// roach performs the replacements and transformations to convert the source
// mysql data for usage with crdb. It will return the output bytes or an error.
func (r *radroach) roach(input []byte) (output []byte, err error) {
	// TODO: Abstract each stage into modular testable units.
	// Stage 1: Syntax.
	output = input
	for pattern, replacement := range simpleReplacements {
		output = pattern.ReplaceAll(output, replacement)
	}

	// Stage 2: Tables.
	// Split the source data down by table.
	tables := tables.FindAll(output, -1)

	tableConstraints := make(map[string][][]byte)
	for i := range tables {
		t := tables[i]

		name := tableName.FindSubmatch(t)
		if name == nil {
			continue
		}
		tableName := string(name[1])

		// Stage 2.1: Extract the constraints from each table definition.
		tableConstraints[tableName] = constraints.FindAll(t, -1)
		t = constraints.ReplaceAll(t, empty)

		// Stage 2.2: Enums.
		if r.enumToCheck {
			tableConstraints[tableName] = enumsToChecks(t, tableConstraints[tableName])
		}
		t = enum.ReplaceAll(t, []byte(" TEXT "))

		// Stage 2.3: Tidy.
		t = blankLines.ReplaceAll(t, empty)
		t = trailingCommas.ReplaceAll(t, []byte("\n);"))

		tables[i] = t
	}

	// Stage 2.4: Rewrite constraints.
	for table, constraints := range tableConstraints {
		for _, constraint := range constraints {
			constraint = bytes.TrimSuffix(constraint, []byte(","))
			constraint = append(constraint, []byte(";")...)

			tableConstraint := append(
				[]byte(fmt.Sprintf("ALTER TABLE %s ADD ", table)),
				constraint...,
			)

			tables = append(tables, tableConstraint)
		}
	}

	output = bytes.Join(tables, []byte("\n"))

	return
}

// log only logs errors if radroach is in verbose mode.
func (r *radroach) log(err error) {
	if !r.verbose {
		return
	}

	r.logger.Println(err)
}

// enumToCheck returns an option modifier to enable enum to check constraints.
func enumToCheck(e bool) option {
	return func(r *radroach) {
		r.enumToCheck = e
	}
}

// verboseLogging returns an option modifier to enable verbose logging.
func verboseLogging(v bool) option {
	return func(r *radroach) {
		r.verbose = v
	}
}

// newRadRoach instantiates a new instance of radroach with the required source
// file, destination file and any options.
func newRadRoach(src, dst string, opts ...option) *radroach {
	rr := &radroach{
		src:    src,
		dst:    dst,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}

	for _, opt := range opts {
		opt(rr)
	}

	return rr
}

// enumsToChecks converts MySQL ENUM fields to PgSQL style constraint checks.
func enumsToChecks(t []byte, input [][]byte) [][]byte {
	// Extract each line containing an enum from the table DDL.
	lines := enumLine.FindAllSubmatch(t, -1)

	for _, line := range lines {
		// Cleanup the line.
		l := bytes.Trim(line[0], ` `)

		// Get the column name for the constraint.
		column := bytes.Trim(bytes.Split(l, space)[0], ` "`)

		// Get the enum values for the constraint.
		values := enum.FindSubmatch(l)
		strValues := string(values[1])

		// Append the constraint.
		input = append(input, []byte(
			fmt.Sprintf(
				"CONSTRAINT check_%[1]s CHECK (%[1]s IN (%[2]s))",
				column,
				strValues,
			),
		))
	}

	return input
}
