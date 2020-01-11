FROM golang:1.13-alpine AS build
COPY . /src
WORKDIR /src
RUN CGO_ENABLED=0 go build -mod=vendor -o=combiner ./cmd/combiner

FROM alpine:3.11
COPY --from=build /src/combiner /
USER 10000:10000
ENTRYPOINT /combiner
