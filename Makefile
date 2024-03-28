test:
	go test -v .

coverage:
	go test -coverprofile=coverage.out -v .

coverage-html: coverage
	go tool cover -html=coverage.out
