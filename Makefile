.SILENT:

build:
	go build

run-sample:
	go run main.go download -u http://localhost:8080/50MB-TESTFILE.ORG.pdf

start-caddy:
	podman compose up -d

stop-caddy:
	podman compose down

download-large-sample:
	curl -o ./data/50MB-TESTFILE.ORG.pdf https://files.testfile.org/PDF/50MB-TESTFILE.ORG.pdf

clean:
	go clean
	if [ -f 50MB-TESTFILE.ORG.pdf ]; then \
		rm 50MB-TESTFILE.ORG.pdf; \
	fi
	if [ -f ./data/50MB-TESTFILE.ORG.pdf ]; then \
		rm ./data/50MB-TESTFILE.ORG.pdf; \
	fi
