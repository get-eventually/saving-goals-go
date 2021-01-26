FROM golang:1.15 AS builder

WORKDIR /go/src/github.com/eventually-rs/saving-goals-go

ADD go.mod go.mod
ADD go.sum go.sum
RUN go mod download

COPY . .

RUN go build ./cmd/saving-goals-api

# ------------------------------------------------------------------------------

FROM ubuntu:bionic

RUN groupadd -r savinggoals && useradd --no-log-init -r -g savinggoals savinggoals
USER savinggoals

COPY --from=builder --chown=savinggoals:savinggoals \
    /go/src/github.com/eventually-rs/saving-goals-go/saving-goals-api \
    /bin/saving-goals-api

ENTRYPOINT ["saving-goals-api"]
