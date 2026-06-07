GOOGLE_CLIENT_ID     ?=
GOOGLE_CLIENT_SECRET ?=

LDFLAGS := -X main.googleClientID=$(GOOGLE_CLIENT_ID) -X main.googleClientSecret=$(GOOGLE_CLIENT_SECRET)

.PHONY: build build-windows build-macos dev test fmt vet generate frontend

build:
	wails build -ldflags "$(LDFLAGS)"

build-windows:
	wails build -nsis -ldflags "$(LDFLAGS)"

build-macos:
	wails build --platform darwin/universal -ldflags "$(LDFLAGS)"

dev:
	wails dev

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

generate:
	wails generate module

frontend:
	cd frontend && npm run build
