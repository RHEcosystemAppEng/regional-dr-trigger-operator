FROM registry.access.redhat.com/ubi9/go-toolset:1.19 as build

USER root
WORKDIR /mcra
COPY . .
RUN make build
USER default

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.2

USER root
COPY --from=build /mcra/build/mcra /usr/bin/mcra
COPY --from=build /mcra/LICENSE /licenses/mcra-license
ENTRYPOINT ["/usr/bin/mcra"]
USER 1001
