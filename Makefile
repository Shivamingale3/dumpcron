BINARY = /usr/local/bin/dumpcron
CONFIG_DIR = /etc/dumpcron
STATE_DIR = /var/lib/dumpcron
PIDEX_DIR = /etc/pidex/custom.d
SYSTEMD_DIR = /etc/systemd/system

build:
	go build -o dumpcron ./cmd/dumpcron

install: build
	install -m 755 dumpcron $(BINARY)
	mkdir -p $(CONFIG_DIR) $(STATE_DIR)
	cp dumpcron.conf $(PIDEX_DIR)/
	cp dumpcron.service $(SYSTEMD_DIR)/
	systemctl daemon-reload

uninstall:
	-systemctl stop dumpcron
	-systemctl disable dumpcron
	rm -f $(BINARY)
	rm -f $(PIDEX_DIR)/dumpcron.conf
	rm -f $(SYSTEMD_DIR)/dumpcron.service
	rm -rf $(STATE_DIR)

validate:
	go build -o dumpcron ./cmd/dumpcron
	./dumpcron validate

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f dumpcron

.PHONY: build install uninstall validate test vet clean
