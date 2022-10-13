test:
	docker build -t kickback-common-tester .
	docker run --rm kickback-common-tester

coverage-report:
	go test -v -coverprofile cover.out ./...
	go tool cover -html=cover.out -o cover.html
	open cover.html