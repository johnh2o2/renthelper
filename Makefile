deployment.zip: main
	zip deployment.zip main

main: vendor
	GOOS=linux go build -o main cmd/renthelper/main.go

vendor:
	dep ensure

.PHONY: clean
clean:
	rm deployment.zip main
	rm -r vendor
