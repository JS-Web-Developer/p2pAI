.PHONY: daemon coordinator proto

daemon:
	cd services/daemon && go build -o bin/p2pai-daemon ./cmd/p2pai-daemon

coordinator:
	cd services/coordinator && go build -o bin/p2pai-coordinator ./cmd/p2pai-coordinator

proto:
	buf generate packages/proto
