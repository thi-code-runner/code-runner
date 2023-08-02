FROM golang:alpine as builder

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	bash \
	ca-certificates

COPY . /go/src/code-runner

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
		make \
	&& cd /go/src/code-runner\
	&& make build\
	&& mv ./bin/code-runner /usr/bin/code-runner \
	&& apk del .build-deps \
	&& rm -rf /go

FROM alpine:latest
EXPOSE 8080

COPY --from=builder /usr/bin/code-runner /usr/bin/code-runner
WORKDIR /usr/bin/

ENTRYPOINT [ "code-runner" ]