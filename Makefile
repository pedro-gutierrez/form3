# Easily download all dependencies
deps:
	@cd cmd; go get -d -v; cd ..  

# Run the app locally, using memory
# storage, enabling the admin apis and exposing prometheus metrics
run-with-sqlite3:
	@go run cmd/main.go --metrics=true --admin=true

# Run the app locally, connecting to a postgres database, enabling the admin apis
# and exposing prometheus metrics
run-with-postgres:
	@go run cmd/main.go --repo=postgres --repo-uri=postgresql://form3:form3@localhost:5432/form3?sslmode=disable --metrics=true --admin=true

# Build a new docker image
docker-build:
	@docker build -t pedrogutierrez/form3:latest .

# Run docker image (uses in memory sqlite3 by default)
docker-run:
	@docker run --name form3 -p 8080:8080 pedrogutierrez/form3:latest

# Stop and remove the docker image
docker-stop:
	@docker stop form3
	@docker rm form3


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
