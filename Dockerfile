ARG ARCH

##
# BUILD CONTAINER
##

FROM alpine:3.13.2 as builder

RUN \
apk add --no-cache ca-certificates

##
# RELEASE CONTAINER
##

FROM ${ARCH}/busybox:1.32-glibc

WORKDIR /

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY tfcw /usr/local/bin/

# Run as nobody user
USER 65534

ENTRYPOINT ["/usr/local/bin/tfcw"]
CMD [""]
