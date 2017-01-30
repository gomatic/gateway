FROM scratch

MAINTAINER nicerobot "https://github.com/gomatic/gateway"

ENV TMP=/
ENV TEMP=/

WORKDIR /

ENV HOME=/
ENV PWD=/
ENV PATH=/

COPY gateway-linux-amd64 /gateway

ENTRYPOINT ["/gateway"]
CMD ["--debug", "--verbose", "--mock"]
