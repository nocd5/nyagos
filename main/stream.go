package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"../history"
	"../readline"
)

func NewCmdStreamConsole() (func(context.Context) (string, error), *history.Container) {
	history1 := new(history.Container)
	editor := readline.Editor{History: history1, Prompt: printPrompt}

	histPath := filepath.Join(AppDataDir(), "nyagos.history")
	history1.Load(histPath)
	history1.Save(histPath)

	return func(ctx context.Context) (string, error) {
		var line string
		var err error
		for {
			line, err = editor.ReadLine(ctx)
			if err != nil {
				return line, err
			}
			var isReplaced bool
			line, isReplaced = history1.Replace(line)
			if isReplaced {
				fmt.Fprintln(os.Stdout, line)
			}
			if line != "" {
				break
			}
		}

		wd, err := os.Getwd()
		if err != nil {
			wd = ""
		}
		row := history.Line{Text: line, Dir: wd, Stamp: time.Now()}
		history1.PushLine(row)
		fd, err := os.OpenFile(histPath, os.O_APPEND, 0600)
		if err != nil && os.IsNotExist(err) {
			fd, err = os.Create(histPath)
		}
		if err == nil {
			fmt.Fprintln(fd, row.String())
			fd.Close()
		} else {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		return line, err
	}, history1
}

func NewCmdStreamFile(r io.Reader) func(ctx context.Context) (string, error) {
	scanner := bufio.NewScanner(r)
	return func(ctx context.Context) (string, error) {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return "", err
			} else {
				return "", io.EOF
			}
		}
		return strings.TrimRight(scanner.Text(), "\r\n"), nil
	}
}