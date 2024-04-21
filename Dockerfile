FROM golang:1.22.2-alpine as build-base

ARG DATABASE_URL

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go test -tags=unit -v ./...

RUN go build -o ./out/assessment .


### ------------

FROM alpine:3.19
COPY --from=build-base /app/out/assessment /app/assessment

ENV DATABASE_URL=postgres://postgres:postgres@localhost:5432/ktaxes?sslmode=disable

CMD ["/app/assessment"]