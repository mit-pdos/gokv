.PHONY: check

check:
	test -z "$$(gofmt -d -s .)"
	go vet -composites=false ./...
	go build ./...

fix:
	gofmt -w -s .
