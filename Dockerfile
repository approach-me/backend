FROM lnogueir/approach-prototools:latest AS protobuilder

RUN git clone https://github.com/approach-me/service-definitions && \
    mkdir -p protos && \
    find service-definitions -name *.proto | xargs protoc --go_out=protos --go-grpc_out=protos

FROM golang:1.18.3 AS builder

WORKDIR /app

COPY . .

COPY --from=protobuilder /protos protos/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go

FROM scratch AS container-env

COPY --from=builder /app/main /

EXPOSE 9090

ENTRYPOINT ["/main"]

# Use --target local-env and --output=protos to generate files locally.
FROM scratch AS local-env

COPY --from=protobuilder /protos ./
