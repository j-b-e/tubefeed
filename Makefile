.PHONY: generate
generate:
	@go generate ./...

.PHONY: build
build: sqlc generate
	 @go build main.go

clean:
	rm -rf internal/sqlc/db.go \
		internal/sqlc/models.go \
		internal/sqlc/query.sql.go
	rm -f main tubefeed

mrproper: clean
	rm -rf config/ audio/

go-mod-update:
	@go get -u ./...
	@go mod tidy

sqlc:
	@go run github.com/sqlc-dev/sqlc/cmd/sqlc compile

air:
	@env DATABASE_PATH=memory air

update-htmx-dep:
	@curl -fsSL "https://cdn.jsdelivr.net/npm/htmx.org/dist/htmx.min.js" -o static/htmx.min.js
	@curl -fsSL "https://cdn.jsdelivr.net/npm/htmx-ext-response-targets/dist/response-targets.min.js" -o static/htmx-ext-response-targets.min.js
