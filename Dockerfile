FROM scratch
ADD conduit /bin/conduit
ENTRYPOINT ["/bin/conduit"]
EXPOSE 8080
CMD ["-h"]
