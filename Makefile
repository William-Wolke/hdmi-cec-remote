SHELL := /bin/bash

GO_BIN = cec-remote
GO_SRC = .
SCRIPT = cec-remote.sh
SERVICE = cec-remote.service
TARGET_HOST = "192.168.0.133"

PREFIX = /usr/local
BIN_DIR = $(PREFIX)/bin
SYSTEMD_DIR = /etc/systemd/system

all: $(GO_BIN)

$(GO_BIN):
	go build -o $(GO_BIN) $(GO_SRC)

build-arm:
	GOOS=linux GOARCH=arm GOARM=7 go build -o $(GO_BIN) $(GO_SRC)

install: $(GO_BIN)
	install -m 0755 $(GO_BIN) $(BIN_DIR)
	install -m 0755 $(SCRIPT) $(BIN_DIR)
	install -m 0644 $(SERVICE) $(SYSTEMD_DIR)
	systemctl daemon-reload
	systemctl enable $(SERVICE)
	systemctl restart $(SERVICE)

clean:
	rm -f $(GO_BIN)

sync-remote:
	make clean
	make build-arm
	scp -r ./* $(TARGET_HOST):~/hdmi-cec-remote/
	ssh -t 192.168.0.133 "cd ~/hdmi-cec-remote/ && sudo make install"

