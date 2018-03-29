# Radroach
A _radical_ tool for converting [MySQL](https://www.mysql.com) database dumps for use with [CockroachDB](https://www.cockroachlabs.com/product/cockroachdb)

> This package is not yet suitable for usage...in any way.

## Installation
Install as you would a normal package:
```bash
go get -u github.com/LUSHDigital/radroach
```

## Usage
Firstly you must have a pre-prepared MySQL dump ready for conversion. Radroach
requires you to have created this using `mysqldump` with some specific options:
```bash
mysqldump -h[host] -u[user] -p [database name] --compatible=postgresql --compact --skip-add-drop-table --skip-add-locks --skip-comments > dump.sql
```
> Opinionated I know but it makes our lives much easier

Then you just need to run `radroach` passing it your source file and the name of
the destination file:
```bash
radroach [FLAGS...] SOURCE_MYSQL_DUMP DESTINATION_CRDB_DUMP
```

### Full Usage
```bash
Usage: radroach [FLAGS...] SOURCE_MYSQL_DUMP DESTINATION_CRDB_DUMP
  -v	verbose logging mode
```

## Roadmap
- [x] Load MySQL dump
- [x] Simple regex replacements for types and syntax
- [ ] Break dump down by table, extract foreign keys
- [ ] Re-write dump with foreign keys after table creation
- [ ] Produce a working SQL dump for CockroachDB
- [ ] Cobra cmd support
- [ ] Test all the things