.PHONY: run
run:
	go run ./main.go -addr $(addr)

.PHONY: test
test:
	GOCACHE=off go test .
