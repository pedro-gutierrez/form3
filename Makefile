# Easily download all dependencies
deps:
	@cd cmd; go get -d -v; cd ..  

# Run the app locally, using disc 
# storage
run:
	@go run cmd/main.go --repo-uri=data/data.db --repo-migrations=./schema 

# Run the app locally, using memory
# storage
run-for-test:
	@go run cmd/main.go --repo-uri=data/test.db --metrics=true --http-logs=true --repo-logs=true --repo-migrations=./schema --admin=true

# Run the app with profiling on
run-for-debug:
	@go run cmd/main.go --metrics=true --repo-uri=data/debug.db --http-logs=true --repo-logs=true --repo-migrations=./schema --admin=true --profiling=true


# Build an executable
build:
	@go build cmd/main.go

# Build a docker image
docker:
	@docker build -t pedrogutierrez/form3:latest .

# Run all BDD scenarios
bdd:
	@cd test; godog; cd ..

# Run individual BDD scenarios
# This target looks for scenarios tagged @wip
bdd-wip:
	@cd test; godog --tags=wip; cd ..

# Start GoDoc's online documentation
# Payments package is described at:
# http://localhost:6060/pkg/github.com/pedro-gutierrez/form3/cmd/payments/
doc:
	@godoc -http=:6060
