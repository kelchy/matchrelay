
golint:
	export GOBIN=$${PWD};\
	go get -u golang.org/x/lint/golint

lint: golint
	./golint -set_exit_status ./...
