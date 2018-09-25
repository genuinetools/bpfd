FROM golang:alpine as builder
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates

COPY . /go/src/github.com/jessfraz/bpfd

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
		make \
	&& cd /go/src/github.com/jessfraz/bpfd \
	&& make static \
	&& mv bpfd /usr/bin/bpfd \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

FROM alpine:latest

COPY --from=builder /usr/bin/bpfd /usr/bin/bpfd
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT [ "bpfd" ]
CMD [ "--help" ]
