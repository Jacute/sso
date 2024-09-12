FROM golang:1.22.4

WORKDIR /sso
ENV CONFIG_PATH=/sso/config/prod.yaml
RUN mkdir storage

COPY . .

RUN go mod tidy
RUN go run ./cmd/migrator/main.go --storage-path=/sso/storage/sso.db --migrations-path=/sso/migrations

CMD [ "go", "run", "cmd/sso/main.go" ]