package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"text/template"

	"github.com/pkg/errors"
)

func main() {
	code, err := run()
	if err != nil {
		log.Fatal("runj: ", err)
	}
	os.Exit(code)
}

func run() (int, error) {
	jsonPath := flag.String("f", "", "json path")
	flag.Parse()

	tmpl, err := template.New("cmd").Parse(flag.Arg(0))
	if err != nil {
		return 0, errors.Wrap(err, "コマンドテンプレートの解釈に失敗しました")
	}

	var w io.Reader = os.Stdin
	if *jsonPath != "" {
		f, err := os.Open(*jsonPath)
		if err != nil {
			return 0, errors.Wrap(err, "ファイルの読み込みに失敗しました")
		}
		defer f.Close()
		w = f
	}

	var vars interface{}
	if err := json.NewDecoder(w).Decode(&vars); err != nil {
		return 0, errors.Wrap(err, "JSONのパースに失敗しました")
	}

	cmd := bytes.NewBuffer(nil)
	if err := tmpl.Execute(cmd, vars); err != nil {
		return 0, errors.Wrap(err, "変数の展開に失敗しました")
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
		return 0, errors.Wrap(err, "コマンドの実行に失敗しました")
	}
	return 0, nil
}
