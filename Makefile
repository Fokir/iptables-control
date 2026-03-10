APP_NAME := system-control
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -s -w -X main.version=$(VERSION)

# === Development ===

.PHONY: dev-backend
dev-backend:
	go run ./cmd/server

.PHONY: dev-frontend
dev-frontend:
	cd web && npm run dev

.PHONY: dev
dev:
	@make -j2 dev-backend dev-frontend

# === Build ===

.PHONY: build-frontend
build-frontend:
	cd web && npm ci && npm run build
	rm -rf cmd/server/frontend
	cp -r web/dist cmd/server/frontend

.PHONY: build
build: build-frontend
	go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME) ./cmd/server

.PHONY: build-linux-amd64
build-linux-amd64: build-frontend
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-linux-amd64 ./cmd/server

.PHONY: build-linux-arm64
build-linux-arm64: build-frontend
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(APP_NAME)-linux-arm64 ./cmd/server

# === Quality ===

.PHONY: lint
lint:
	golangci-lint run ./...
	cd web && npm run lint

.PHONY: test
test:
	go test -race -cover ./...

# === Clean ===

.PHONY: clean
clean:
	rm -rf bin/ web/dist/ cmd/server/frontend/
	mkdir -p cmd/server/frontend
	echo '<!DOCTYPE html><html><body>placeholder</body></html>' > cmd/server/frontend/index.html
