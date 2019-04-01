# Easily download all dependencies
deps:
	@cd cmd; go get -d -v; cd ..  

# Run the app locally, using disc 
# storage
run:
	@go run cmd/main.go --repo-uri=data/data.db --repo-migrations=./schema 

# Run the app locally, using memory
# storage
run-with-sqlite3:
	@go run cmd/main.go --repo=sqlite3 --repo-uri=data/sqlite3/test.db --metrics=true --http-logs=true --repo-migrations=./schema --admin=true

# Run the app locally, using memory
# storage
run-with-postgres:
	@go run cmd/main.go --repo=postgres --repo-uri=postgresql://form3:form3@localhost:5432/form3?sslmode=disable --metrics=true --http-logs=true --repo-migrations=./schema --admin=true

# Run the app with profiling on
run-for-debug:
	@go run cmd/main.go --metrics=true --repo-uri=data/debug.db --http-logs=true --repo-migrations=./schema --admin=true --profiling=true


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


# Stop and remove postgres
postgres-stop:
	@docker stop postgres
	@docker rm postgres

# Start a new postgres db from docker
# Pre-existing data is removed, and a new form3 user and database
# are created from scratch
postgres-start:
	@rm -rf $(PWD)/data/postgres
	@mkdir -p $(PWD)/data/postgres/data $(PWD)/data/postgres/run $(PWD)/data/postgres/init
	@echo "CREATE USER form3 WITH PASSWORD 'form3'; CREATE DATABASE form3; GRANT ALL ON DATABASE form3 to form3;" > $(PWD)/data/postgres/init/form3.sql
	@docker run --name postgres -e POSTGRES_PASSWORD=docker -p 5432:5432 -v "$(PWD)/data/postgres/data:/var/lib/postgresql/data" -v "$(PWD)/data/postgres/run:/var/run/postgresql" -v "$(PWD)/data/postgres/init:/docker-entrypoint-initdb.d/" -d postgres

# Connect to our postgres form3 database
postgres-connect:
	@docker exec -it postgres psql -h localhost -U form3 -d form3
