.PHONY: deploy
deploy: 
	go build ${bin/infra} build/main.go
	@./main

.PHONY: build 
deploy: 
	go build ${bin/infra} build/main.go
