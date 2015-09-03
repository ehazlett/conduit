FROM alpine
RUN apk add -U git bash py-pip
RUN pip install -U docker-compose
COPY conduit /bin/conduit
COPY ./certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/bin/conduit"]
EXPOSE 8080
CMD ["-h"]
