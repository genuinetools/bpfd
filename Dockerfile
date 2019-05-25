FROM debian:stretch-slim as builder
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

# Add non-free apt sources
RUN sed -i "s#deb http://deb.debian.org/debian buster main#deb http://deb.debian.org/debian buster main contrib non-free#g" /etc/apt/sources.list

RUN apt-get update && apt-get install -y \
    ca-certificates \
	clang \
	curl \
	gcc \
	git \
	g++ \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Install dependencies for libbcc
# FROM: https://github.com/iovisor/bcc/blob/master/INSTALL.md#install-build-dependencies
RUN apt-get update && apt-get install -y \
	debhelper \
	cmake \
	libllvm3.9 \
	llvm-dev \
	libclang-dev \
	libelf-dev \
	bison \
	flex \
	libedit-dev \
	clang-format \
	python \
	python-netaddr \
	python-pyroute2 \
	luajit \
	libluajit-5.1-dev \
	arping \
	iperf \
	ethtool \
	devscripts \
	zlib1g-dev \
	libfl-dev \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Build libbcc
ENV BCC_VERSION v0.9.0
RUN git clone --depth 1 --branch "$BCC_VERSION" https://github.com/iovisor/bcc.git /usr/src/bcc
WORKDIR /usr/src/bcc
RUN mkdir build \
	&& cd build \
	&& cmake .. -DCMAKE_INSTALL_PREFIX=/usr \
	&& make \
	&& make install \
	&& make clean \
	&& make

# Install Go
ENV GO_VERSION 1.11
RUN curl -fsSL "https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz" \
	| tar -xzC /usr/local

COPY . /go/src/github.com/genuinetools/bpfd

WORKDIR /go/src/github.com/genuinetools/bpfd
RUN make \
	&& mv bpfd /usr/bin/bpfd

FROM debian:stretch-slim

# Add non-free apt sources
RUN sed -i "s#deb http://deb.debian.org/debian buster main#deb http://deb.debian.org/debian buster main contrib non-free#g" /etc/apt/sources.list
RUN apt-get update && apt-get install -y \
	bison \
    ca-certificates \
	clang \
	cmake \
	flex \
	git \
	libclang-dev \
	libelf-dev \
	libluajit-5.1-dev \
	llvm-dev \
	make \
    --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/bin/bpfd /usr/bin/bpfd
COPY --from=builder /usr/src/bcc /usr/src/bcc
COPY examples /etc/bpfd/

RUN cd /usr/src/bcc/build && make install

ENTRYPOINT [ "bpfd" ]
CMD [ "--help" ]
