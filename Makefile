
GO_BIN = cec-program
GO_SRC = .
SCRIPT = cec-remote.sh
SERVICE = cec-program.service

PREFIX = /usr/local
BIN_DIR = $(PREFIX)/bin
SYSTEMD_DIR = /etc/systemd/system

all: $(GO_BIN)

$(GO_BIN):
	go build -o $(GO_BIN) $(GO_SRC)

install: $(GO_BIN)
	install -m 0755 $(GO_BIN) $(BIN_DIR)
	install -m 0755 $(SCRIPT) $(BIN_DIR)
	install -m 0644 $(SERVICE) $(SYSTEMD_DIR)
	systemctl daemon-reload
	systemctl enable $(SERVICE)
	systemctl restart $(SERVICE)

clean:
	rm -f $(GO_BIN)
