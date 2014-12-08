FROM golang:1.3.3-onbuild
EXPOSE 8080
ENTRYPOINT ["/go/src/app/conduit"]
CMD ["-h"]
