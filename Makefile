build:
	go build -o radroach-dev

test: build
	./radroach-dev input.sql output.sql

enum_test: build
	./radroach-dev -enum-to-check input.sql output.sql