FROM golang:1.24-alpine as builder
WORKDIR /go/src/app
COPY . .
#RUN apt update && apt upgrade -y
RUN CGO_ENABLED=0 go build

FROM alpine:latest as final
WORKDIR /srv/
RUN mkdir /srv/dota_patch_bot
RUN apk add libc6-compat
#RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
COPY --from=builder /go/src/app/dota_patch_bot /srv/dota_patch_bot/

CMD ["/srv/dota_patch_bot/dota_patch_bot"]
