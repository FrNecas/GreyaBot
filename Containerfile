FROM golang:alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

COPY . /go/src/github.com/FrNecas/GreyaBot

WORKDIR /go/src/github.com/FrNecas/GreyaBot

RUN apk --no-cache add ca-certificates

RUN apk add --no-cache git \
    && go get . \
    && apk del git

RUN go build main.go

CMD ["./main"]

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/FrNecas/GreyaBot/main /
COPY --from=builder /go/src/github.com/FrNecas/GreyaBot/db/migrations /db/migrations

USER 1001

ENTRYPOINT ["/main"]
