#!/bin/bash

# Check if any files in internal/handlers have changed since the last Swagger generation
if find internal/handlers -type f -name '*.go' -newer docs/swagger.json | grep .; then
  echo "Detected changes in route handlers. Running swag init..."
  swag init -g ./cmd/server/main.go
else
  echo "No route changes. Skipping swag init."
fi

# Proceed with the build after checking for changes
echo "Running go build..."
go build -o ./tmp/main ./cmd/server
