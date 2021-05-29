FROM alpine:latest

COPY locker /locker

EXPOSE 80

ENTRYPOINT [ "/locker" ]
