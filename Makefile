
test:
	docker build -t kickback-common-tester .
	docker run --rm kickback-common-tester
