ARG GOLANG_VERSION=1.21.3

FROM golang:${GOLANG_VERSION} as base
ENV CGO_ENABLED=0 GO111MODULE=on
WORKDIR /go/src/github.com/smoynes/elsie

FROM base as builder
ADD . .
RUN --mount=type=cache,target=/go/cache \
    go env -w GOCACHE=/go/cache/build  &&\
    go env -w GOMODCACHE=/go/cache/mod && \
    go env && \
    go install -v .

FROM base as dev
WORKDIR /home/elsie

ARG UID=1000
RUN adduser --uid=${UID} --disabled-password elsie
USER elsie:elsie

COPY --from=builder /go/bin/elsie /usr/local/bin/elsie
COPY --from=builder /go/src/github.com/smoynes/elsie/docs \
    /home/elsie

CMD ["/usr/local/bin/elsie", "demo"]
