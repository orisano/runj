package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/template"

	"github.com/pkg/errors"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("runj: ")

	code, err := run()
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

func run() (int, error) {
	jsonPath := flag.String("f", "", "json path")
	flag.Parse()

	tmpl, err := template.New("cmd").Parse(strings.Join(flag.Args(), " "))
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse command template")
	}

	var w io.Reader = os.Stdin
	if *jsonPath != "" {
		f, err := os.Open(*jsonPath)
		if err != nil {
			return 0, errors.Wrap(err, "failed to open file")
		}
		defer f.Close()
		w = f
	}

	var vars interface{}
	if err := json.NewDecoder(w).Decode(&vars); err != nil {
		return 0, errors.Wrap(err, "failed to parse JSON")
	}

	cmd := bytes.NewBuffer(nil)
	if err := tmpl.Execute(cmd, vars); err != nil {
		return 0, errors.Wrap(err, "failed to expand variables")
	}

	c := exec.Command("/bin/bash", "-c", cmd.String())
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			if ws, ok := ee.Sys().(syscall.WaitStatus); ok {
				return ws.ExitStatus(), nil
			}
		}
		return 0, errors.Wrap(err, "failed to execute command")
	}
	return 0, nil
}
