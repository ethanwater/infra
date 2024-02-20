.PHONY: deploy
deploy: 
	@./main

.PHONY: build 
build:  
	go build ${bin/infra} build/main.go
	@./main
