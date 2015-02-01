FROM scratch
COPY conduit /bin/conduit
EXPOSE 8080
ENTRYPOINT ["/bin/conduit"]
CMD ["-h"]
