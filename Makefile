build:
	npm install --production --no-audit
	templ generate view
	go generate ./cmd
	go build -o bin/main ./cmd
templ:
	templ generate -watch -proxy=http://localhost:6969
