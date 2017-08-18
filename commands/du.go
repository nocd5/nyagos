package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zetamatta/nyagos/shell"
)

func du_(path string, output func(string, int64), blocksize int64) (int64, error) {
	fd, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	stat, err := fd.Stat()
	if err != nil {
		return 0, err
	}
	if !stat.IsDir() {
		fd.Close()
		return ((stat.Size() + blocksize - 1) / blocksize) * blocksize, nil
	}
	files, err := fd.Readdir(-1)
	if err != nil {
		fd.Close()
		return 0, err
	}
	var diskuse int64 = 0
	dirs := make([]string, 0, len(files))
	for _, file1 := range files {
		if file1.IsDir() {
			dirs = append(dirs, file1.Name())
		} else {
			diskuse += ((file1.Size() + blocksize - 1) / blocksize) * blocksize
		}
	}
	if err := fd.Close(); err != nil {
		return diskuse, err
	}
	for _, dir1 := range dirs {
		fullpath := filepath.Join(path, dir1)
		diskuse1, err := du_(fullpath, output, blocksize)
		if err != nil {
			return diskuse, err
		}
		output(fullpath, diskuse1)
		diskuse += diskuse1
	}
	return diskuse, nil
}

func cmd_du(_ context.Context, cmd *shell.Cmd) (int, error) {
	output := func(name string, size int64) {
		fmt.Fprintf(cmd.Stdout, "%d\t%s\n", size/1024, name)
	}
	count := 0
	for _, arg1 := range cmd.Args[1:] {
		if arg1 == "-s" {
			output = func(_ string, _ int64) {}
			continue
		}
		size, err := du_(arg1, output, 4096)
		count++
		if err != nil {
			fmt.Fprintf(cmd.Stderr, "%s: %s\n", arg1, err)
			continue
		}
		fmt.Fprintf(cmd.Stdout, "%d\t%s\n", size/1024, arg1)
	}
	if count <= 0 {
		size, err := du_(".", output, 4096)
		if err != nil {
			return 1, err
		}
		fmt.Fprintf(cmd.Stdout, "%d\t%s\n", size/1024, ".")
	}
	return 0, nil
}
