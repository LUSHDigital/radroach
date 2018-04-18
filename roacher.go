package main

import (
	"bytes"
	"fmt"
	"regexp"
)

var (
	empty = []byte("")

	simpleReplacements = map[*regexp.Regexp][]byte{
		regexp.MustCompile("`"):                                         empty,
		regexp.MustCompile(`^.* ENGINE=.*$/\)`):                         empty,
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
		regexp.MustCompile(` enum\((.*)\) `):                            []byte(" TEXT "),
		regexp.MustCompile("  KEY "):                                    []byte("  INDEX "),
		regexp.MustCompile("  FULLTEXT KEY "):                           []byte("  INDEX "),
		regexp.MustCompile("ON UPDATE CURRENT_TIMESTAMP"):               empty,
		regexp.MustCompile(" ON DELETE CASCADE"):                        empty,
		regexp.MustCompile("DEFAULT b'0',$"):                            []byte("DEFAULT 0,"),
		regexp.MustCompile("DEFAULT b'1',$"):                            []byte("DEFAULT 1,"),
		regexp.MustCompile(`int DEFAULT '(.*)',$`):                      []byte("INT DEFAULT $1,"),
		regexp.MustCompile(`int NOT NULL DEFAULT '(.*)',$`):             []byte("INT NOT NULL DEFAULT $1,"),
		regexp.MustCompile(`\(decimal(.*)\) NOT NULL DEFAULT '(.*)',$`): []byte("$1 NOT NULL DEFAULT $2,"),
		regexp.MustCompile(`\(decimal(.*)\) DEFAULT '(.*)',$`):          []byte("$1 DEFAULT $2,"),
		regexp.MustCompile("  UNIQUE KEY "):                             []byte("  UNIQUE INDEX "),
	}

	blankLines     = regexp.MustCompile(`\s\s\n`)
	constraints    = regexp.MustCompile("(?smU)CONSTRAINT.*REFERENCES.*$")
	tables         = regexp.MustCompile("(?smU)^(CREATE|INSERT).*;$")
	tableName      = regexp.MustCompile(`CREATE TABLE "(.*)"`)
	trailingCommas = regexp.MustCompile(`,\n\);`)
)

type roacher struct {
	sourceData []byte
}

func newRoacher(data []byte) *roacher {
	return &roacher{data}
}

func (r *roacher) roach() (output []byte, err error) {
	output = r.sourceData
	for pattern, replacement := range simpleReplacements {
		output = pattern.ReplaceAll(output, replacement)
	}

	// Split the source data down by table.
	tables := tables.FindAll(output, -1)

	tableConstraints := make(map[string][][]byte)
	for i := range tables {
		t := &tables[i]

		name := tableName.FindSubmatch(*t)
		if name == nil {
			continue
		}

		// Extract the constraints from each table definition.
		tableConstraints[string(name[1])] = constraints.FindAll(*t, -1)
		*t = constraints.ReplaceAll(*t, empty)

		// Tidy.
		*t = blankLines.ReplaceAll(*t, empty)
		*t = trailingCommas.ReplaceAll(*t, []byte("\n);"))
	}

	for table, constraints := range tableConstraints {
		for _, constraint := range constraints {
			constraint = bytes.Replace(constraint, []byte(","), empty, -1)
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
