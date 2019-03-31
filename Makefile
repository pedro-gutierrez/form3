# Easily download all dependencies
deps:
	@cd cmd; go get -d -v; cd ..  

# Run the app locally, using disc 
# storage
disc:
	@go run cmd/main.go --dbUri=data/data.db 

# Run the app locally, using memory
# storage
run-for-test:
	@go run cmd/main.go --metrics=true --repo-uri=data/data.db --http-logs=true --repo-logs=true --repo-migrations=./schema --admin=true --profiling=true

# Build an executable
build:
	@go build cmd/main.go

# Build a docker image
docker:
	@docker build -t pedrogutierrez/form3:latest .

# Run BDDs
bdd-all:
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
