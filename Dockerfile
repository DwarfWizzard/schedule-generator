FROM golang:1.25.1-trixie AS build

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download 

COPY . .

RUN go build -o svc ./cmd/schedule-generator/.

FROM debian

WORKDIR /application

COPY --from=build /build/svc /application/svc

RUN ls -la

CMD ["./svc"]