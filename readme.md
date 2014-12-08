# Conduit
Experimental deployment system for Docker.

Conduit exposes an endpoint that receives webhooks (i.e. from Docker Hub).  Upon receiving the hook, Conduit will pull the new image, deploy a new container from the updated image and then remove the original container. 

# Usage
Docker.

```
docker run
    -d
    --name conduit
    -v /var/run/docker.sock:/var/run/docker.sock
    ehazlett/conduit -r <repo-name> -t <token>
```

Where `<repo-name>` is a Docker repository name such as `ehazlett/go-demo` and `<token>` is a custom token string.  The `-r` arg can be specified multiple times.

Example:

```
docker run
    -d
    --name conduit
    -v /var/run/docker.sock:/var/run/docker.sock
    ehazlett/conduit -r ehazlett/go-demo -t s3cr3+
```
Then add a webhook url to `http://<your-conduit-host>:<your-conduit-port>?token=<token>`
