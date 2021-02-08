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

RUN go get github.com/go-delve/delve/cmd/dlv
RUN go build -gcflags="all=-N -l" -o main

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/FrNecas/GreyaBot/main /
COPY --from=builder /go/src/github.com/FrNecas/GreyaBot/db/migrations /db/migrations
COPY --from=builder /go/bin/dlv /

USER 1001

ENTRYPOINT ["/dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/main"]
