run:
	go build -o radroach-dev
	./radroach-dev --enum-to-check input.sql output.sql
	rm radroach-dev