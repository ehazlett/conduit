FROM scratch
ADD conduit /bin/conduit
ENTRYPOINT ["/bin/conduit"]
CMD ["-h"]
