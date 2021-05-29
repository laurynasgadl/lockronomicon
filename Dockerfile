FROM alpine:latest

COPY lockronomicon /lockronomicon

EXPOSE 80

ENTRYPOINT [ "/lockronomicon" ]
