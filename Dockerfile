FROM golang:1.23 AS build
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /out/event-fanout ./cmd/server

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/event-fanout /event-fanout
EXPOSE 8081
ENTRYPOINT ["/event-fanout"]
