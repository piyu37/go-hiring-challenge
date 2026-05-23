COVERAGE_UNIT := coverage.unit.out
COVERAGE_INTEGRATION := coverage.integration.out
COVERAGE_MERGED := coverage.out
COVERPKG := github.com/mytheresa/go-hiring-challenge/...

tidy ::
	@go mod tidy && go mod vendor

seed ::
	@go run cmd/seed/main.go

run ::
	@go run cmd/server/main.go

test ::
	@go test -v -count=1 -race ./... -coverprofile=$(COVERAGE_UNIT) -covermode=atomic

integration-test ::
	@go test -v -count=1 -tags=integration ./models/... \
		-coverprofile=$(COVERAGE_INTEGRATION) -covermode=atomic \
		-coverpkg=$(COVERPKG)

coverage ::
	@$(MAKE) test
	@echo ""
	@echo "=== Unit test coverage ==="
	@go tool cover -func=$(COVERAGE_UNIT)
	@echo ""
	@echo "=== Total unit coverage ==="
	@go tool cover -func=$(COVERAGE_UNIT) | tail -1

test-all ::
	@$(MAKE) test
	@$(MAKE) integration-test
	@go run github.com/wadey/gocovmerge@latest \
		$(COVERAGE_UNIT) $(COVERAGE_INTEGRATION) > $(COVERAGE_MERGED)
	@echo ""
	@echo "=== Combined unit + integration coverage ==="
	@go tool cover -func=$(COVERAGE_MERGED)
	@echo ""
	@echo "=== Total combined coverage ==="
	@go tool cover -func=$(COVERAGE_MERGED) | tail -1

docker-up ::
	docker compose up -d

docker-down ::
	docker compose down
