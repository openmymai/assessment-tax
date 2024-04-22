FROM golang:1.22 as build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /tax

FROM build-stage AS run-test-stage
RUN go test -tags=unit -v ./...

### ------------

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /tax /tax

EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/tax"]