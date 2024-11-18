postgres://vadymivanenko@localhost:5432/identity?sslmode=disable

migrate create -ext sql -dir migrations/ tokens

migrate -source file://migrations -database "postgres://vadymivanenko@localhost:5432/identity?sslmode=disable" down 1