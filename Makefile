build:
	templ generate view
	go build -o bin/main ./cmd
templ:
	templ generate -watch -proxy=http://localhost:8080
