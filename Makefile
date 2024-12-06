.PHONY: check

check:
	test -z "$$(gofmt -d -s .)"
	go vet -composites=false ./...
	go build ./...
	go test ./...

fix:
	./update-grackle.py
	gofmt -w -s .
