APP_NAME=logviewer
SRC=$(wildcard *.go)

.PHONY: all deps build install clean run

all: build install

deps:
	@echo "Tidying up Go dependencies..."
	go mod tidy

build: deps
	@echo "Building $(APP_NAME)..."
	go build -o $(APP_NAME) $(SRC)

install: build
	@echo "Installing $(APP_NAME) to /usr/local/bin..."
	sudo mv $(APP_NAME) /usr/local/bin/

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(APP_NAME)
	@if [ -f /usr/local/bin/$(APP_NAME) ]; then \
		echo "Removing $(APP_NAME) from /usr/local/bin..."; \
		sudo rm -f /usr/local/bin/$(APP_NAME); \
	fi

run: build
	@echo "Running $(APP_NAME)..."
	./$(APP_NAME)
