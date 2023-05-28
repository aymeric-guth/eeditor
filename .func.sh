build() {
	# go build .
	go build -ldflags "-s -w" .
}
run() {
	go run .
}
