package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/template"
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
		return 0, fmt.Errorf("parse command template: %w", err)
	}

	var r io.Reader = os.Stdin
	if *jsonPath != "" {
		f, err := os.Open(*jsonPath)
		if err != nil {
			return 0, fmt.Errorf("open json: %w", err)
		}
		defer f.Close()
		r = f
	}

	var vars interface{}
	if err := json.NewDecoder(r).Decode(&vars); err != nil {
		return 0, fmt.Errorf("parse json: %w", err)
	}

	cmd := &bytes.Buffer{}
	if err := tmpl.Execute(cmd, vars); err != nil {
		return 0, fmt.Errorf("expand variables: %w", err)
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
		return 0, fmt.Errorf("execute command: %w", err)
	}
	return 0, nil
}
