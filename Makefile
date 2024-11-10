# Define directories
PROTO_DIR := proto
PROTO_SRC := $(PROTO_DIR)/content
CMD_DIR := cmd
INTERNAL_DIR := internal

# Install necessary tools for protobuf compilation
install-tools:
	@echo "Installing necessary tools..."
	# Install protoc-gen-go and protoc-gen-go-grpc for protobuf and gRPC support
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "Tools installed successfully."

# Compile proto files
generate-proto:
	@echo "Generating gRPC code for ContentService..."
	PATH=$(HOME)/go/bin:$(PATH) protoc -I=$(PROTO_DIR) \
		--go_out=. \
		--go-grpc_out=. \
		$(PROTO_DIR)/**/*.proto

# Download dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "Dependencies updated."

# Run the application
run-app:
	@echo "Running ContentService application..."
	go run $(CMD_DIR)/main.go

# Run the entire pipeline
all: install-tools generate-proto tidy run-app
