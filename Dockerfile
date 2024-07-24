FROM golang:1.22 AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux go build .

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=build /app/sypchal .

EXPOSE 3000

# Run
CMD ["./sypchal"]