FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/smm-api .
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/smm-api /smm-api
EXPOSE 8080
ENTRYPOINT ["/smm-api"]
