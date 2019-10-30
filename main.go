package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
)

type SternEntry struct {
	Message       string `json:"message"`
	Namespace     string `json:"namespace"`
	PodName       string `json:"podName"`
	ContainerName string `json:"containerName"`
}

func checkStdinPipe() {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// this is ok
	} else {
		panic("stdin: nothing being piped in ,_,")
	}
}

func getFileHandle(fileCache map[string]*os.File, name string) *os.File {
	entry, ok := fileCache[name]
	if !ok {
		f, err := os.OpenFile(name, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(fmt.Sprintf("openFile: %s: %s", name, err))
		}
		fmt.Fprintf(os.Stderr, "+%s\n", name)
		fileCache[name] = f
		return f
	} else {
		return entry
	}
}

func process(entry SternEntry, fileCache map[string]*os.File) {
	f := getFileHandle(fileCache, entry.PodName)
	_, fWriteErr := f.WriteString(entry.Message)
	if fWriteErr != nil {
		panic(fmt.Sprintf("writeback: err: %s, filename=[%s]", fWriteErr, entry.PodName))
	}
}

func cleanup(fileCache map[string]*os.File) {
	for name, f := range fileCache {
		fmt.Fprintf(os.Stderr, "-%s\n", name)
		f.Close()
	}
}

func main() {
	checkStdinPipe()
	fileCache := make(map[string]*os.File)
	defer cleanup(fileCache)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			cleanup(fileCache)
		}
	}()
	reader := bufio.NewReader(os.Stdin)
	decoder := json.NewDecoder(reader)
	for {
		var entry SternEntry
		err := decoder.Decode(&entry)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(fmt.Sprintf("json: decode: %s", err))
		}
		process(entry, fileCache)
	}
}
