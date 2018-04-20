package main

import (
	"bytes"
	"fmt"
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

type roacher struct {
	sourceData []byte
}

func newRoacher(data []byte) *roacher {
	return &roacher{data}
}

func (r *roacher) roach() (output []byte, err error) {
	// TODO: Abstract each stage into modular testable units.
	// Stage 1: Syntax.
	output = r.sourceData
	for pattern, replacement := range simpleReplacements {
		output = pattern.ReplaceAll(output, replacement)
	}

	// Stage 2: Constraints.
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

		processEnums(string(name[1]), t, tableConstraints)

		// Tidy.
		*t = blankLines.ReplaceAll(*t, empty)
		*t = trailingCommas.ReplaceAll(*t, []byte("\n);"))
	}

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

func processEnums(table string, t *[]byte, c map[string][][]byte) {
	// If enum-to-check has been requested, replace enums with check constraints,
	// otherwise, just replace enum identifiers with text identifiers.
	if opts.enumToCheck {
		lines := enumLine.FindAllSubmatch(*t, -1)

		for _, line := range lines {
			// Cleanup the line.
			l := bytes.Trim(line[0], ` `)

			// Get the column name for the constraint.
			column := bytes.Trim(bytes.Split(l, space)[0], ` "`)

			// Get the enum values for the constraint.
			values := enum.FindSubmatch(l)
			strValues := string(values[1])

			// Append the constraint.
			c[table] = append(c[table], []byte(fmt.Sprintf(
				"CONSTRAINT check_%[1]s CHECK (%[1]s IN (%[2]s))",
				column,
				strValues)))
		}
	}

	*t = enum.ReplaceAll(*t, []byte(" TEXT "))
}
