FROM r.j3ss.co/bcc
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

# Install Go
ENV GO_VERSION 1.11
RUN curl -fsSL "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz" \
	| tar -xzC /usr/local

# Install google/protobuf
ENV PROTOBUF_VERSION v3.6.1
RUN set -x \
	&& export PROTOBUF_PATH="$(mktemp -d)" \
	&& curl -fsSL "https://github.com/google/protobuf/archive/${PROTOBUF_VERSION}.tar.gz" \
		| tar -xzC "$PROTOBUF_PATH" --strip-components=1 \
	&& ( \
		cd "$PROTOBUF_PATH" \
		&& ./autogen.sh \
		&& ./configure --prefix=/usr/local \
		&& make \
		&& make install \
		&& ldconfig \
	) \
	&& rm -rf "$PROTOBUFPATH"

# Install Go deps
RUN go get golang.org/x/lint/golint
RUN go get honnef.co/go/tools/cmd/staticcheck
RUN go get github.com/golang/protobuf/proto
RUN go get github.com/golang/protobuf/protoc-gen-go

COPY . /go/src/github.com/genuinetools/bpfd

WORKDIR /go/src/github.com/genuinetools/bpfd
ENTRYPOINT ["sh", "-c"]
