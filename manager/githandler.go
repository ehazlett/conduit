package manager

import (
	"net/http"
	"net/http/cgi"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	log "github.com/Sirupsen/logrus"
)

var (
	postReceiveHookTemplate = `#!/bin/bash
REPO_DIR="{{.RepoDir}}"
/bin/bash {{.RepoDir}}/hooks/deploy
`
	deployTemplate = `
#!/bin/bash
NAME={{.Name}}
REPO_DIR="{{.RepoDir}}"
WORK_DIR="{{.WorkDir}}"

echo " --> Deploying $NAME"

unset GIT_INDEX_FILE
git --work-tree=$WORK_DIR --git-dir=$REPO_DIR checkout -f

cd $WORK_DIR

OUT=$(docker-compose up -d 2>&1)

exec < /dev/tty

IFS=$'\n' read -rd '' -a COMPOSE_OUT <<<"$OUT"
for LINE in "${COMPOSE_OUT[@]}"; do
    echo " --> $LINE"
done

echo " --> Deploy for $NAME complete"
`
)

type (
	HookConfig struct {
		Name    string
		WorkDir string
		RepoDir string
	}
)

func createRepository(repoDir string) error {
	log.Infof("creating new repository: dir=%s", repoDir)

	cmd := exec.Command("git", "--bare", "init", repoDir)
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func setupPostReceiveHook(name, repoDir, workDir string) error {
	log.Debugf("creating post-receive hook: name=%s repo=%s", name, repoDir)

	hookPath := filepath.Join(repoDir, "hooks", "post-receive")
	// hook
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		if err := os.Remove(hookPath); err != nil {
			return err
		}
	}
	fc, err := os.Create(hookPath)
	if err != nil {
		return err
	}
	fc.Close()

	// make executable
	if err := os.Chmod(hookPath, 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(hookPath, os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	deployPath := filepath.Join(repoDir, "hooks", "deploy")
	// hook
	if _, err := os.Stat(deployPath); os.IsNotExist(err) {
		if err := os.Remove(deployPath); err != nil {
			return err
		}
	}

	dc, err := os.Create(deployPath)
	if err != nil {
		return err
	}
	dc.Close()

	// make executable
	if err := os.Chmod(deployPath, 0755); err != nil {
		return err
	}

	df, err := os.OpenFile(deployPath, os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer df.Close()

	hookConfig := &HookConfig{
		Name:    name,
		WorkDir: workDir,
		RepoDir: repoDir,
	}

	t := template.New("post-receive-hook")
	tmpl, err := t.Parse(postReceiveHookTemplate)
	if err != nil {
		log.Errorf("error parsing hook template: %s", err)
		return err
	}

	if err := tmpl.Execute(f, hookConfig); err != nil {
		log.Errorf("error executing hook template: %s", err)
		return err
	}

	dt := template.New("deploy")
	dTmpl, err := dt.Parse(deployTemplate)
	if err != nil {
		log.Errorf("error parsing deploy template: %s", err)
		return err
	}

	if err := dTmpl.Execute(df, hookConfig); err != nil {
		log.Errorf("error executing deploy template: %s", err)
		return err
	}

	return nil
}

func (m *Manager) setupWorkDir(repoDir, repoWorkDir string) error {
	log.Debugf("setting up repo work dir: repodir=%s repoworkdir=%s",
		repoDir,
		repoWorkDir,
	)

	// remove existing dir if present
	if _, err := os.Stat(repoWorkDir); err == nil {
		log.Debugf("removing existing work dir: path=%s", repoWorkDir)
		if err := os.RemoveAll(repoWorkDir); err != nil {
			return err
		}
	}

	// clone to work dir for deployment
	cmd := exec.Command("git", "clone", repoDir, repoWorkDir)
	if out, err := cmd.Output(); err != nil {
		log.Error(out)
		return err
	}

	return nil
}

func (m *Manager) gitHandler(w http.ResponseWriter, r *http.Request) {
	username := "git"
	path := r.URL.Path
	pathParts := strings.Split(path, "/")

	repoDir := filepath.Join(m.repoRootDir, filepath.Join(pathParts[1], pathParts[2]))
	repoName := filepath.Base(repoDir)

	// create repo if not exists
	if _, err := os.Stat(repoDir); os.IsNotExist(err) {
		if err := createRepository(repoDir); err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	repoWorkDir := filepath.Join(m.repoWorkDir, repoName)
	if err := m.setupWorkDir(repoDir, repoWorkDir); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ensure post-receive hook exists and is updated
	if err := setupPostReceiveHook(
		repoName,
		repoDir,
		repoWorkDir,
	); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h := &cgi.Handler{
		Path: "/bin/sh",
		Args: []string{"-c", "git http-backend"},
		Dir:  ".",
		Env: []string{
			"GIT_PROJECT_ROOT=" + m.repoRootDir,
			"GIT_HTTP_EXPORT_ALL=1",
			"REMOTE_USER=" + username,
		},
	}

	remoteIP := r.Header.Get("X-Forwarded-For")
	if remoteIP == "" {
		remoteIP = r.RemoteAddr
	}

	if filepath.Base(r.URL.Path) == "git-receive-pack" {
		log.Infof("deploy: name=%s ip=%s", repoName, remoteIP)
	}

	log.Debugf("%s path=%s", r.Method, r.URL.Path)

	h.ServeHTTP(w, r)
}
