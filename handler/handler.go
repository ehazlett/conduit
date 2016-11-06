package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"
	"github.com/ehazlett/conduit/types"
	"github.com/ehazlett/conduit/version"
	"github.com/gorilla/mux"
)

type HandlerConfig struct {
	ListenAddr   string
	Repositories []string
	Token        string
}

type info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Handler struct {
	config *HandlerConfig
	client client.APIClient
}

func New(cfg *HandlerConfig) (*Handler, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	return &Handler{
		config: cfg,
		client: cli,
	}, nil
}

func (h *Handler) info(w http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(w).Encode(info{
		Name:    version.Name(),
		Version: version.Version(),
	}); err != nil {
		rErr := fmt.Errorf("error getting application info")
		logrus.Error(rErr)
		http.Error(w, rErr.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) handleHook(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	var hook *types.Webhook
	if err := json.NewDecoder(r.Body).Decode(&hook); err != nil {
		logrus.Errorf("error decoding webhook: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	repoName := hook.Repository.RepositoryName

	logrus.WithFields(logrus.Fields{
		"timestamp": time.Now(),
		"name":      repoName,
	}).Debug("webhook received")

	responsePayload := &types.CallbackPayload{
		TargetURL: "",
	}

	if token != h.config.Token {
		rErr := fmt.Errorf("invalid token %s", token)
		responsePayload.State = "error"
		responsePayload.Description = rErr.Error()

		if hook.CallbackURL != "" {
			if err := h.sendResponse(responsePayload, hook.CallbackURL); err != nil {
				logrus.Error(err)
			}
		}

		http.Error(w, rErr.Error(), http.StatusUnauthorized)
		logrus.Error(rErr)

		return

	}

	if !h.isValidRepository(repoName) {
		rErr := fmt.Errorf("%s is not in whitelist", repoName)

		responsePayload.State = "error"
		responsePayload.Description = rErr.Error()
		if hook.CallbackURL != "" {
			if err := h.sendResponse(responsePayload, hook.CallbackURL); err != nil {
				logrus.Error(err)
			}
		}

		http.Error(w, rErr.Error(), http.StatusUnauthorized)
		logrus.Error(rErr)

		return
	}

	logrus.Debugf("deploying %s", repoName)

	// TODO: deploy
	if err := h.deploy(repoName); err != nil {
		rErr := fmt.Errorf("error deploying %s: %s", repoName, err)

		responsePayload.State = "error"
		responsePayload.Description = rErr.Error()
		if hook.CallbackURL != "" {
			if err := h.sendResponse(responsePayload, hook.CallbackURL); err != nil {
				logrus.Error(err)
			}
		}

		http.Error(w, rErr.Error(), http.StatusUnauthorized)
		logrus.Error(rErr)

		return
	}

	responsePayload.State = "success"
	responsePayload.Description = fmt.Sprintf("conduit deployed %s", repoName)
	w.WriteHeader(http.StatusOK)

	if hook.CallbackURL != "" {
		if err := h.sendResponse(responsePayload, hook.CallbackURL); err != nil {
			logrus.Error(err)
		}
	}
}

func (h *Handler) Run() error {
	r := mux.NewRouter()

	r.HandleFunc("/", h.info).Methods("GET")
	r.HandleFunc("/", h.handleHook).Methods("POST")

	http.Handle("/", r)

	logrus.Infof("%s listening on %s", version.Name(), h.config.ListenAddr)
	logrus.Infof("repositories: %s", strings.Join(h.config.Repositories, ", "))

	// TODO: TLS
	return http.ListenAndServe(h.config.ListenAddr, nil)
}
