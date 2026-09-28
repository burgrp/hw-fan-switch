# Optional aliases for the local development inventory. Deployment consumers
# invoke the same BleRiot commands from their own authoritative inventory.
.PHONY: build flash gdb rtt install-pack disassembly new hub

build:
	go run ./cmd/dev node build --name fan

flash:
	go run ./cmd/dev node build --name fan --flash --rtt

disassembly:
	go run ./cmd/dev node build --name fan --disassembly

gdb rtt install-pack:
	go run ./cmd/dev node $@ --name fan

new:
	go run ./cmd/dev node new

hub:
	go run ./cmd/dev hub --diagnostics rf
