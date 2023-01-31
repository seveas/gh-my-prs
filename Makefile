gh-my-prs: *.go go.*
	go build

lint:
	golangci-lint run

.PHONY: lint
