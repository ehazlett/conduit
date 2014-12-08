package manager

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ehazlett/conduit/hub"
	"github.com/gorilla/mux"
	"github.com/samalba/dockerclient"
	log "github.com/sirupsen/logrus"
)

type (
	Manager struct {
		repoWhitelist []string
		dockerUrl     string
		authUsername  string
		authPassword  string
		authEmail     string
	}

	HookResponse struct {
		Message string `json:"message,omitempty"`
	}

	Info struct{}
)

func NewManager(repoWhitelist []string, dockerUrl string, authUsername string, authPassword string, authEmail string, debug bool) (*Manager, error) {
	if debug {
		log.SetLevel(log.DebugLevel)
	}

	return &Manager{
		repoWhitelist: repoWhitelist,
		dockerUrl:     dockerUrl,
		authUsername:  authUsername,
		authPassword:  authPassword,
		authEmail:     authEmail,
	}, nil
}

func (m *Manager) index(w http.ResponseWriter, r *http.Request) {
	resp := Info{}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (m *Manager) receive(w http.ResponseWriter, r *http.Request) {
	data := &hub.Webhook{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repoName := data.Repository.RepoName

	if !m.isValidRepo(repoName) {
		http.Error(w, fmt.Sprintf("%s not on whitelist", repoName), http.StatusBadRequest)
		return
	}

	if err := m.deploy(repoName); err != nil {
		http.Error(w, fmt.Sprintf("error deploying: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (m *Manager) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/", m.index).Methods("GET")
	r.HandleFunc("/", m.receive).Methods("POST")
	http.Handle("/", r)

	log.Infof("starting conduit")
	log.Infof("repos: %v", m.repoWhitelist)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("unable to start: %s", err)
	}
}

func (m *Manager) isValidRepo(repo string) bool {
	valid := false
	for _, v := range m.repoWhitelist {
		if v == "*" {
			valid = true
			break
		}
		if v == repo {
			valid = true
			break
		}
	}

	return valid
}

func (m *Manager) authConfig() *dockerclient.AuthConfig {
	if m.authUsername == "" || m.authPassword == "" || m.authEmail == "" {
		return nil
	}

	return &dockerclient.AuthConfig{
		Username: m.authUsername,
		Password: m.authPassword,
		Email:    m.authEmail,
	}
}

func (m *Manager) deploy(repo string) error {
	docker, err := dockerclient.NewDockerClient(m.dockerUrl, nil)
	if err != nil {
		return err
	}

	containers, err := docker.ListContainers(false, false, "")
	if err != nil {
		return err
	}

	authConfig := m.authConfig()

	for _, c := range containers {
		if strings.Contains(c.Image, repo) {
			cId := c.Id[:10]
			log.Infof("deploying new image for container: %s", cId)

			log.Debugf("%s: pulling new image", cId)
			docker.PullImage(repo, authConfig)

			log.Debugf("%s: launching new container", cId)

			cfg, err := docker.InspectContainer(c.Id)
			if err != nil {
				return err
			}

			id, err := docker.CreateContainer(cfg.Config, "")
			if err != nil {
				return err
			}

			if err := docker.StartContainer(id, cfg.HostConfig); err != nil {
				return err
			}

			log.Debugf("%s: stopping old container", cId)
			if err := docker.StopContainer(c.Id, 5); err != nil {
				return err
			}

			log.Debugf("%s: removing old container", cId)
			if err := docker.RemoveContainer(c.Id, true); err != nil {
				return err
			}

			log.Infof("%s: deployed new container", cId)
		}
	}

	return nil
}
