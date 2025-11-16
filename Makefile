SHELL := /bin/bash

GO_BIN = cec-remote
GO_SRC = .
SCRIPT = cec-remote.sh
SERVICE = cec-remote.service
TARGET_HOST = "192.168.0.133"
TARGET_USER = "william"

PREFIX = /usr/local
BIN_DIR = $(PREFIX)/bin
SYSTEMD_DIR = /home/$(TARGET_USER)/.config/systemd/user

all: $(GO_BIN)

$(GO_BIN):
	go build -o $(GO_BIN) $(GO_SRC)

build-arm:
	GOOS=linux GOARCH=arm GOARM=7 go build -o $(GO_BIN) $(GO_SRC)

install: $(GO_BIN)
	install -m 0755 $(GO_BIN) $(BIN_DIR)
	install -m 0755 $(SCRIPT) $(BIN_DIR)
	install -m 0644 $(SERVICE) $(SYSTEMD_DIR)
	# Run in a user session:
	# systemctl --user daemon-reload
	# systemctl --user enable $(SERVICE)
	# systemctl --user restart $(SERVICE)

reload-service:
	systemctl --user daemon-reload
	systemctl --user enable $(SERVICE)
	systemctl --user restart $(SERVICE)

clean:
	rm -f $(GO_BIN)

sync-remote:
	make clean
	make build-arm
	scp -r ./* $(TARGET_USER)@$(TARGET_HOST):/home/$(TARGET_USER)/hdmi-cec-remote/
	ssh -t $(TARGET_USER)@$(TARGET_HOST) "cd /home/$(TARGET_USER)/hdmi-cec-remote/ && sudo make install"

