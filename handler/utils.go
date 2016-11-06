package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/ehazlett/conduit/types"
)

func (h *Handler) isValidRepository(repo string) bool {
	if len(h.config.Repositories) == 0 {
		return false
	}

	for _, r := range h.config.Repositories {
		if repo == r {
			return true
		}
	}

	return false
}

func (h *Handler) sendResponse(payload *types.CallbackPayload, callbackURL string) error {
	logrus.Debugf("sending response payload: callback=%s", callbackURL)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}

	if _, err := http.Post(callbackURL, "application/json", &buf); err != nil {
		return err
	}

	return nil
}

func (h *Handler) deploy(repo string) error {
	logrus.WithFields(logrus.Fields{
		"name": repo,
	}).Info("deploying")

	containers, err := h.client.ContainerList(context.Background(), dockertypes.ContainerListOptions{
		Size: false,
		All:  false,
	})
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"name":      repo,
		"instances": len(containers),
	}).Debugf("checking containers for repository")

	for _, c := range containers {
		image := c.Image

		logrus.WithFields(logrus.Fields{
			"repo":  repo,
			"image": image,
		}).Debugf("checking image for repo")

		if image != repo {
			logrus.WithFields(logrus.Fields{
				"image": image,
				"repo":  repo,
			}).Error("container image does not match repo")
			continue
		}

		logrus.WithFields(logrus.Fields{
			"image": image,
		}).Debug("deploying")

		cID := c.ID[:10]
		logrus.WithFields(logrus.Fields{
			"container": cID,
		}).Info("deploying new image for container")

		logrus.WithFields(logrus.Fields{
			"container": cID,
			"image":     image,
		}).Debug("pulling new image for container")
		if _, err := h.client.ImagePull(context.Background(), image, dockertypes.ImagePullOptions{}); err != nil {
			return err
		}

		logrus.WithFields(logrus.Fields{
			"container": cID,
		}).Debug("creating new container")

		cfg, err := h.client.ContainerInspect(context.Background(), c.ID)
		if err != nil {
			return err
		}

		// reset hostname to get new id
		cfg.Config.Hostname = ""

		resp, err := h.client.ContainerCreate(context.Background(), cfg.Config, cfg.HostConfig, nil, "")
		if err != nil {
			return err
		}

		portBinds := false

		if len(cfg.HostConfig.PortBindings) > 0 {
			portBinds = true
		}
		// check for port bindings; if exist, stop/remove container first
		// so new container can bind to specified ports; otherwise
		// allow the container to start first and allocate random ports
		if portBinds {
			if err := h.removeContainer(c.ID); err != nil {
				return err
			}
		}

		if err := h.client.ContainerStart(context.Background(), resp.ID, dockertypes.ContainerStartOptions{}); err != nil {
			return err
		}

		if !portBinds {
			if err := h.removeContainer(c.ID); err != nil {
				return err
			}
		}

		logrus.WithFields(logrus.Fields{
			"container": resp.ID[:10],
		}).Info("started new container")
	}

	return nil
}

func (h *Handler) removeContainer(id string) error {
	cID := id[:10]

	logrus.WithFields(logrus.Fields{
		"container": cID,
	}).Debug("stopping container")
	timeout := time.Second * 5
	if err := h.client.ContainerStop(context.Background(), id, &timeout); err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"container": cID,
	}).Debug("removing container")
	if err := h.client.ContainerRemove(context.Background(), id, dockertypes.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		return err
	}

	return nil
}
