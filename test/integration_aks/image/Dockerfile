FROM ubuntu:18.04

COPY "./azurelustreplugin" "/app/azurelustreplugin"
COPY "./run_integration_test.sh" "/app/run_integration_test.sh"

RUN chmod +x "/app/run_integration_test.sh"

RUN apt-get update

# Fix vulnerability CVE-2018-7169
RUN apt-get upgrade -y login passwd

COPY --from=golang /usr/local/go/ /usr/local/go/

ENV GOPATH="/root/go"
ENV PATH="/usr/local/go/bin:/root/go/bin:${PATH}"

WORKDIR "/app"

LABEL maintainers="fuxiuyin;joe-atzinger;Justin-Jin"
LABEL description="Azure Lustre CSI driver integration test"

ENTRYPOINT ["/app/run_integration_test.sh"]
