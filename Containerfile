# TODO replace with ubi9/go-toolset once go 1.20 is supported
FROM docker.io/library/golang:1.20 as builder

WORKDIR /mcra
COPY . .
RUN make build

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.2

USER root
COPY --from=builder /mcra/build/mcra /usr/bin/mcra
COPY --from=builder /mcra/LICENSE /licenses/mcra-license
ENTRYPOINT ["/usr/bin/mcra"]
USER 1001
