FROM registry.access.redhat.com/ubi9/go-toolset:1.20 as builder

USER root
WORKDIR /mcra
COPY . .
RUN make build
USER default

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.3

USER root
COPY --from=builder /mcra/build/mcra /usr/bin/mcra
COPY --from=builder /mcra/LICENSE /licenses/mcra-license
ENTRYPOINT ["/usr/bin/mcra"]
USER 1001
