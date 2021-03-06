# Easily download all dependencies
deps:
	@cd cmd; go get -d -v; cd ..  

# Run the app locally, using memory
# storage, enabling the admin apis and exposing prometheus metrics
sqlite3: deps
	@go run cmd/main.go --metrics=true --admin=true

# Build a new docker image
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
# Form3 specific stuff is at:
# http://localhost:6060/pkg/github.com/pedro-gutierrez/form3/
doc:
	@godoc -http=:6060
