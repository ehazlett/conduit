# Conduit
Experimental deployment system for Docker.

Conduit exposes an endpoint that receives webhooks (i.e. from Docker Hub).  Upon receiving the hook, Conduit will pull the new image, deploy a new container from the updated image and then remove the original container. 

# Usage
Docker.

```
docker run
    -d
    --name conduit
    -p 8080:8080
    -v /var/run/docker.sock:/var/run/docker.sock
    ehazlett/conduit -r <repo-name> -t <token>
```

Where `<repo-name>` is a Docker repository name such as `ehazlett/go-demo` and `<token>` is a custom token string.  The `-r` arg can be specified multiple times.

Example:

```
docker run
    -d
    --name conduit
    -p 8080:8080
    -v /var/run/docker.sock:/var/run/docker.sock
    ehazlett/conduit -r ehazlett/go-demo -t s3cr3+
```
Then add a webhook url to `http://<your-conduit-host>:<your-conduit-port>?token=<token>`

You can also specify a list of tags for deploy.  Conduit will only deploy
and rotate containers that are using that tag.  For example, if you have
containers with both `v1` and `v2` tags running, if you specify `v2` as a tag
in Conduit, it will only deploy the `v2` containers when receiving a webhook.

# Testing
To simulate a webhook using curl:

```
curl -d '{"repository": {"repo_name": "namespace/reponame"}}' http://<docker-host-ip>:8080?token=yourtoken
```
