
build:
	go build

run-sample:
	go run main.go download -u http://localhost:8080/sample-50-MB-pdf-file.pdf

start-caddy:
	podman compose up -d

stop-caddy:
	podman compose down

clean:
	go clean
	rm sample-50-MB-pdf-file.pdf
