include .env
export

build:
	@go build -o bin/app ./cmd/app/

run: build
	@./bin/app
	
clean: 
	@rm -rf bin


DB_CONNECTION := $(DATABASE)

drop: 
	@migrate -database "$(DB_CONNECTION)" -path db/migrations down 

mig: 
	@migrate -database "$(DB_CONNECTION)" -path db/migrations up 


