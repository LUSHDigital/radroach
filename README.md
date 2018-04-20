# Radroach
A _radical_ tool for converting [MySQL](https://www.mysql.com) database dumps for use with [CockroachDB](https://www.cockroachlabs.com/product/cockroachdb)

## Installation
Install as you would a normal package:
```bash
go get -u github.com/LUSHDigital/radroach
```

## Usage
Firstly you must have a pre-prepared MySQL dump ready for conversion. Radroach
requires you to have created this using `mysqldump` with some specific options:
```bash
mysqldump -h[host] -u[user] -p[database name] --compatible=postgresql --compact --skip-add-drop-table --skip-add-locks --skip-comments > dump.sql
```
> Opinionated I know but it makes our lives much easier

To generate a MySQL dump for a MySQL table hosted in Docker, use the following steps:

* Jump into the MySQL Docker container shell:

``` bash
docker exec -it <CONTAINER> /bin/bash
```

* Run `mysqldump` as outlined above.

* Copy the file/contents of dump.sql and use that as the input to `radroach`.

Then you just need to run `radroach` passing it your source file and the name of
the destination file:
```bash
radroach [FLAGS...] SOURCE_MYSQL_DUMP DESTINATION_CRDB_DUMP
```

### Full Usage
```bash
Usage: radroach [FLAGS...] SOURCE_MYSQL_DUMP DESTINATION_CRDB_DUMP
  -enum-to-check
    	convert enums to check constraints
  -verbose
    	verbose logging mode
```

## Roadmap
- [x] Load MySQL dump
- [x] Simple regex replacements for types and syntax
- [x] Break dump down by table, extract foreign keys
- [x] Re-write dump with foreign keys after table creation
- [x] Produce a working SQL dump for CockroachDB
- [ ] Refactor codebase for readability and testability
- [ ] Test all the things
- [ ] Cobra cmd support