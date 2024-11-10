postgres://vadymivanenko@localhost:5432/identity?sslmode=disable

migrate create -ext sql -dir migrations/ confirmation

migrate -source file://migrations -database "postgres://vadymivanenko@localhost:5432/identity?sslmode=disable" down 1