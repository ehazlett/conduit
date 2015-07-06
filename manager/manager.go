package manager

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/ehazlett/conduit/hub"
	"github.com/ehazlett/conduit/tlsutils"
	"github.com/ehazlett/conduit/version"
	"github.com/gorilla/mux"
	"github.com/samalba/dockerclient"
)

type (
	Manager struct {
		listenAddr      string
		repoWhitelist   []string
		tags            []string
		dockerUrl       string
		tlsCaCert       string
		tlsCert         string
		tlsKey          string
		allowInsecure   bool
		token           string
		authUsername    string
		authPassword    string
		authEmail       string
		repoRootDir     string
		repoWorkDir     string
		serverTLSCACert string
		serverTLSCert   string
		serverTLSKey    string
	}

	ManagerConfig struct {
		ListenAddr      string
		RepoWhitelist   []string
		Tags            []string
		DockerURL       string
		TLSCACert       string
		TLSCert         string
		TLSKey          string
		AllowInsecure   bool
		AuthUsername    string
		AuthPassword    string
		AuthEmail       string
		Token           string
		RepoRootDir     string
		RepoWorkDir     string
		ServerTLSCACert string
		ServerTLSCert   string
		ServerTLSKey    string
		Debug           bool
	}

	HookResponse struct {
		Message string `json:"message,omitempty"`
	}

	Info struct {
		Name    string `json:"name,omitempty"`
		Version string `json:"version,omitempty"`
	}
)

var (
	AppInfo = Info{
		Name:    "conduit",
		Version: version.FullVersion(),
	}
)

func NewManager(cfg *ManagerConfig) (*Manager, error) {
	if cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	// make root work dir
	if err := os.MkdirAll(cfg.RepoWorkDir, 0755); err != nil {
		return nil, err
	}

	return &Manager{
		listenAddr:      cfg.ListenAddr,
		repoWhitelist:   cfg.RepoWhitelist,
		tags:            cfg.Tags,
		dockerUrl:       cfg.DockerURL,
		tlsCaCert:       cfg.TLSCACert,
		tlsCert:         cfg.TLSCert,
		tlsKey:          cfg.TLSKey,
		allowInsecure:   cfg.AllowInsecure,
		authUsername:    cfg.AuthUsername,
		authPassword:    cfg.AuthPassword,
		authEmail:       cfg.AuthEmail,
		token:           cfg.Token,
		repoRootDir:     cfg.RepoRootDir,
		repoWorkDir:     cfg.RepoWorkDir,
		serverTLSCACert: cfg.ServerTLSCACert,
		serverTLSCert:   cfg.ServerTLSCert,
		serverTLSKey:    cfg.ServerTLSKey,
	}, nil
}

func (m *Manager) index(w http.ResponseWriter, r *http.Request) {
	resp := AppInfo
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (m *Manager) receive(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.Header.Get("X-Forwarded-For")
	if remoteIP == "" {
		remoteIP = r.RemoteAddr
	}

	token := r.Header.Get("X-Token")
	if token != m.token {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		log.Warnf("invalid token: token=%q ip=%s", token, remoteIP)
		return
	}

	data := &hub.Webhook{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	repoName := data.Repository.RepoName

	log.Debugf("webhook received: name=%s", repoName)

	responsePayload := &hub.CallbackPayload{
		TargetUrl: "",
	}

	if !m.isValidRepo(repoName) {
		log.Errorf("invalid repo: name=%s", repoName)
		responsePayload.State = "error"
		responsePayload.Description = fmt.Sprintf("%s is not on whitelist", repoName)
		w.WriteHeader(http.StatusBadRequest)
		if data.CallbackUrl != "" {
			if err := m.sendResponse(responsePayload, data.CallbackUrl); err != nil {
				log.Error(err)
			}
		}
		return
	}

	if err := m.deploy(repoName); err != nil {
		log.Errorf("error deploying: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		responsePayload.State = "error"
		responsePayload.Description = fmt.Sprintf("error deploying: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		if data.CallbackUrl != "" {
			if err := m.sendResponse(responsePayload, data.CallbackUrl); err != nil {
				log.Error(err)
			}
		}
		return
	}

	log.Infof("deployed: repo=%s", repoName)

	// post back to hub
	responsePayload.State = "success"
	responsePayload.Description = fmt.Sprintf("deployed %s", repoName)
	w.WriteHeader(http.StatusOK)

	if data.CallbackUrl != "" {
		if err := m.sendResponse(responsePayload, data.CallbackUrl); err != nil {
			log.Error(err)
		}
	}
}

func (m *Manager) Run() error {
	r := mux.NewRouter()

	r.HandleFunc("/info", m.index).Methods("GET")
	r.HandleFunc("/hook", m.receive).Methods("POST")
	r.HandleFunc("/{repo:.*}", http.HandlerFunc(m.gitHandler))
	http.Handle("/", r)

	log.Infof("conduit v%s", version.FullVersion())
	log.Infof("repos: %v", m.repoWhitelist)
	log.Infof("tags: %v", m.tags)
	log.Infof("repo dir: %s", m.repoRootDir)

	// TLS
	if m.serverTLSCert != "" && m.serverTLSKey != "" {
		log.Info("TLS enabled")
		log.Debugf("tls info: cert=%s key=%s",
			m.serverTLSCert,
			m.serverTLSKey,
		)

		// setup TLS config
		var caCert []byte
		if m.serverTLSCACert != "" {
			ca, err := ioutil.ReadFile(m.serverTLSCACert)
			if err != nil {
				return err
			}

			caCert = ca
		}

		serverCert, err := ioutil.ReadFile(m.serverTLSCert)
		if err != nil {
			return err
		}

		serverKey, err := ioutil.ReadFile(m.serverTLSKey)
		if err != nil {
			return err
		}

		clientAuthType := tls.RequireAndVerifyClientCert
		tlsConfig, err := tlsutils.GetServerTLSConfig(caCert, serverCert, serverKey, m.allowInsecure, clientAuthType)
		if err != nil {
			return err
		}

		s := &http.Server{
			Addr:      m.listenAddr,
			TLSConfig: tlsConfig,
		}

		if err := s.ListenAndServeTLS(m.serverTLSCert, m.serverTLSKey); err != nil {
			return err
		}
	} else {
		if err := http.ListenAndServe(m.listenAddr, nil); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) sendResponse(payload *hub.CallbackPayload, callbackUrl string) error {
	log.Debugf("sending response payload: callback=%s", callbackUrl)

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}

	if _, err := http.Post(callbackUrl, "application/json", &buf); err != nil {
		return err
	}

	return nil
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

func (m *Manager) validTag(tag string) bool {
	for _, t := range m.tags {
		if tag == t {
			return true
		}
	}

	return false
}

func (m *Manager) deploy(repo string) error {
	docker, err := GetDockerClient(
		m.dockerUrl,
		m.tlsCaCert,
		m.tlsCert,
		m.tlsKey,
		m.allowInsecure,
	)
	if err != nil {
		return err
	}

	containers, err := docker.ListContainers(false, false, "")
	if err != nil {
		return err
	}

	authConfig := m.authConfig()

	for _, c := range containers {
		img := strings.Split(c.Image, ":")
		var (
			image string
			tag   string
		)

		if len(img) == 1 {
			image = c.Image
			tag = "latest"
		} else {
			image = strings.Join(img[:len(img)-1], "")
			tag = img[len(img)-1]
		}

		if image == repo && m.validTag(tag) {
			log.Debugf("deploying: image=%s tag=%s", image, tag)
			cId := c.Id[:10]
			log.Infof("deploying new image for container: %s", cId)

			imgName := fmt.Sprintf("%s:%s", repo, tag)
			log.Debugf("%s: pulling new image: repo=%s", cId, imgName)
			docker.PullImage(imgName, authConfig)

			log.Debugf("%s: launching new container", cId)

			cfg, err := docker.InspectContainer(c.Id)
			if err != nil {
				return err
			}

			// reset hostname to get new id
			cfg.Config.Hostname = ""

			id, err := docker.CreateContainer(cfg.Config, "")
			if err != nil {
				return err
			}

			portBinds := false

			if len(cfg.HostConfig.PortBindings) > 0 {
				portBinds = true
			}
			// check for port bindings; if exist, stop/remove container first
			if portBinds {
				if err := m.removeContainer(c.Id); err != nil {
					return err
				}
			}

			if err := docker.StartContainer(id, cfg.HostConfig); err != nil {
				return err
			}

			if !portBinds {
				if err := m.removeContainer(c.Id); err != nil {
					return err
				}
			}

			log.Infof("%s: deployed new container", cId)
		}
	}

	return nil
}

func (m *Manager) removeContainer(id string) error {
	cId := id[:10]
	docker, err := GetDockerClient(
		m.dockerUrl,
		m.tlsCaCert,
		m.tlsCert,
		m.tlsKey,
		m.allowInsecure,
	)
	if err != nil {
		return err
	}

	log.Debugf("%s: stopping old container", cId)
	if err := docker.StopContainer(id, 5); err != nil {
		return err
	}

	log.Debugf("%s: removing old container", cId)
	if err := docker.RemoveContainer(id, true, true); err != nil {
		return err
	}

	return nil
}
