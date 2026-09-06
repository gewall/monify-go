# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS build
ENV GOTOOLCHAIN=local CGO_ENABLED=0
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -trimpath -buildvcs=false -ldflags="-s -w" -o /out/monify ./cmd/monify

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/monify /monify
EXPOSE 8080
ENTRYPOINT ["/monify"]
CMD []
