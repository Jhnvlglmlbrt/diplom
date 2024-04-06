include .env
export

build:
	@go build -o bin/app ./cmd/app/

run: build
	@./bin/app
	
clean: 
	@rm -rf bin

pull:
	@go build -o bin/puller cmd/puller/main.go
	@./bin/puller


DB_CONNECTION := $(DATABASE)

drop: 
	@migrate -database "$(DB_CONNECTION)" -path db/migrations down 

mig: 
	@migrate -database "$(DB_CONNECTION)" -path db/migrations up 


