build: build/ping-monitor.linux build/ping-monitor.macos

build/ping-monitor.linux: main.go go.mod go.sum
	@mkdir -p build
	GOOS=linux go build -o build/ping-monitor.linux

build/ping-monitor.macos: main.go go.mod go.sum
	@mkdir -p build
	GOOS=darwin go build -o build/ping-monitor.macos
