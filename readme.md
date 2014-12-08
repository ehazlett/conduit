# Conduit
Experimental deployment system for Docker.

Conduit exposes an endpoint that receives webhooks (i.e. from Docker Hub).  Upon receiving the hook, Conduit will pull the new image, deploy a new container from the updated image and then remove the original container. 

