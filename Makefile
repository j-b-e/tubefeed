generate:
	@go generate tubefeed/internal/sqlc

build:
	 @go build main.go

clean:
	rm -rf internal/sqlc/db.go \
		internal/sqlc/models.go \
		internal/sqlc/query.sql.go
	rm -f main tubefeed

mrproper: clean
	rm -rf config/ audio/
