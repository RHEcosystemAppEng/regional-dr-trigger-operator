FROM registry.access.redhat.com/ubi10/go-toolset:1.23 as builder

USER root
WORKDIR /rdrtrigger
COPY . .
# using a custom build directory to avoid contaminating local 'build' one
RUN make build LOCALBUILD=internal_build
USER default

FROM registry.access.redhat.com/ubi10/ubi-minimal:10.0

USER root
COPY --from=builder /rdrtrigger/internal_build/rdrtrigger /usr/bin/rdrtrigger
COPY --from=builder /rdrtrigger/LICENSE /licenses/rdrtrigger-license
ENTRYPOINT ["/usr/bin/rdrtrigger"]
USER 1001
