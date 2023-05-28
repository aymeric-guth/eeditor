package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
	"mvdan.cc/sh/expand"
	"mvdan.cc/sh/syntax"
)

type Editor struct {
	Name string
	Path []string
	Env  map[string]string
}

type EditorYaml struct {
	Name string            `yaml:"name"`
	Path interface{}       `yaml:"path"`
	Env  map[string]string `yaml:"env,omitempty"`
}

type Candidate struct {
	Name string
	Path string
	Env  []string
}

func Expand(s string, env func(string) string) (string, error) {
	p := syntax.NewParser()
	word, err := p.Document(strings.NewReader(s))
	if err != nil {
		return "", err
	}
	if env == nil {
		env = os.Getenv
	}
	cfg := &expand.Config{Env: expand.FuncEnviron(env)}
	return expand.Document(cfg, word)
}

func main() {
	var editors []Editor
	var buff []EditorYaml
	var candidates []Candidate
	var env []string
	var configs []string
	var path string
	var config string

	log.SetOutput(ioutil.Discard)
	// Config loading
	path = os.Getenv("EEDITOR_CONFIG")
	if path != "" {
		path, err := Expand(path, nil)
		if err == nil {
			configs = append(configs, path)
		}
	}
	path = os.Getenv("XDG_CONFIG_HOME")
	if path != "" {
		path, err := Expand(path, nil)
		if err == nil {
			configs = append(configs, path+"/eeditor/eeditor.yml")
		}
	}
	configs = append(
		configs,
		os.Getenv("HOME")+"/.config/eeditor/eeditor.yml",
		os.Getenv("HOME")+"/.eeditor.yml",
		"/etc/eeditor/eeditor.yml",
	)
	for _, path := range configs {
		_, err := os.Lstat(path)
		if err == nil {
			config = path
			break
		}
	}
	if config == "" {
		log.Fatalf("Could not find config file")
	}
	data, err := os.ReadFile(config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal([]byte(data), &buff)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	// Pre-processin Optional[string|sequence[string]] -> []string
	for idx, entry := range buff {
		editor := Editor{Name: entry.Name, Path: []string{}, Env: entry.Env}
		switch i := entry.Path.(type) {
		case string:
			editor.Path = append(editor.Path, i)
		case []string:
			editor.Path = append(editor.Path, i...)
		case nil:
			val, err := exec.LookPath(entry.Name)
			if err != nil {
				log.Printf("Could not find %s in PATH", entry.Name)
			} else {
				editor.Path = append(editor.Path, filepath.Dir(val))
				editor.Name = filepath.Base(val)
			}
		default:
			log.Fatalf("Unknown type i=%s idx=%d", i, idx)
		}
		editors = append(editors, editor)
	}

	for _, editor := range editors {
		env = []string{}
		// Environment expansion environment varialbes
		for k, v := range editor.Env {
			val, err := Expand(v, nil)
			if err != nil {
				log.Printf("Could not resolve env=%s, error: %v", val, err)
			} else {
				env = append(env, k+"="+val)
			}
		}
		// Environment expansion for path
		for idx, v := range editor.Path {
			path, err := Expand(v, nil)
			if err != nil {
				log.Printf("Could not resolve env=%s, error: %v", v, err)
			} else {
				editor.Path[idx] = path
			}
		}
		// Dépendances pour 1 entry
		// Environment, besoin de toutes les variables d'environnement
		// Path, besoin d'au moins 1 Path
		// Make Candidate
		// Dès qu'un Candidat est disponible, on peut procéder à la suite de la validation
		// Les candidats sont évalués dans l'ordre de priorité défini dans la config
		for _, v := range editor.Path {
			candidates = append(candidates, Candidate{
				Path: filepath.Join(v, editor.Name),
				Env:  env,
			})
		}
	}

	for _, candidate := range candidates {
		fi, err := os.Lstat(candidate.Path)
		if err != nil {
			log.Printf("Could not stat: %s err=%s", candidate, err)
			continue
		}
		log.Printf("candidate=%s", candidate.Path)
		if fi.Mode().Perm()&0o111 == 0 {
			log.Printf("%s is not executable", candidate)
			continue
		}
		log.Printf("Running %s", candidate)
		cmd := exec.Command(candidate.Path, os.Args[1:]...)
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, candidate.Env...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
		// syscall.Exec not working on OSX
		// syscall.Exec(filepath.Join(editor.Path, editor.Name), os.Args[1:], env)
	}
	log.Fatalf("Could not find any editor")

}
