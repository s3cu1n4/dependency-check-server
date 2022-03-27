FROM owasp/dependency-check:latest

USER root


WORKDIR /src


RUN  mkdir /src/report
RUN  mkdir /src/conf


COPY  conf/conf.yaml /src/conf/
COPY  build/linux_server /src/



ENTRYPOINT ["/src/linux_server"]