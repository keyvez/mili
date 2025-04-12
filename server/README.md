# Go HTMX Demo

A simple Go server demonstrating HTMX integration with Tailwind CSS.

## Prerequisites

- Go 1.x or higher
- Make (optional, for using Makefile commands)

## Getting Started

### Run the server

```bash
make run
```
or without Make:
```bash
go run main.go
```

### Build the application

```bash
make build
```
or without Make:
```bash
go build -o bin/server
```

### Run tests

```bash
make test
```
or without Make:
```bash
go test ./...
```

### Clean build artifacts

```bash
make clean
```
or without Make:
```bash
rm -rf bin/
```

## Project Structure

```
.
├── main.go          # Main server file
├── templates/       # HTML templates
│   └── index.html
├── static/         # Static assets (CSS, JS, images)
└── Makefile        # Build automation
```

## Access the Application

Once running, access the application at:
- http://localhost:8080

## Technologies Used

- Go
- HTMX
- Tailwind CSS
``` 