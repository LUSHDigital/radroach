package main

import "regexp"

var simpleReplacements = map[*regexp.Regexp][]byte{
	regexp.MustCompile("`"):                                         []byte(""),
	regexp.MustCompile(`^.* ENGINE=.*$/\)`):                         []byte(""),
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
	regexp.MustCompile("ON UPDATE CURRENT_TIMESTAMP"):               []byte(""),
	regexp.MustCompile("ON DELETE CASCADE"):                         []byte(""),
	regexp.MustCompile("DEFAULT b'0',$"):                            []byte("DEFAULT 0,"),
	regexp.MustCompile("DEFAULT b'1',$"):                            []byte("DEFAULT 1,"),
	regexp.MustCompile(`int DEFAULT '(.*)',$`):                      []byte("INT DEFAULT $1,"),
	regexp.MustCompile(`int NOT NULL DEFAULT '(.*)',$`):             []byte("INT NOT NULL DEFAULT $1,"),
	regexp.MustCompile(`\(decimal(.*)\) NOT NULL DEFAULT '(.*)',$`): []byte("$1 NOT NULL DEFAULT $2,"),
	regexp.MustCompile(`\(decimal(.*)\) DEFAULT '(.*)',$`):          []byte("$1 DEFAULT $2,"),
	regexp.MustCompile("  UNIQUE KEY "):                             []byte("  UNIQUE INDEX "),
}

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

	return
}
