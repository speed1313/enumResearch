package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"sync/atomic"
)

func main() {
	// Make HTTP GET request
	enumCount := uint64(0)
	var wg sync.WaitGroup

	response, err := http.Get("https://index.golang.org/index")
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()
	scanner := bufio.NewScanner(response.Body)
	type Message struct {
		Path, Version, Timestamp string
	}
	pkgLists := make([]string, 0)
	for scanner.Scan() {
		// json dump
		dec := json.NewDecoder(strings.NewReader(scanner.Text()))
		var m Message
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		pkgLists = append(pkgLists, m.Path)
	}
	for _, pkgname := range pkgLists[0:10] {

		hashDir := sha256.Sum256([]byte(pkgname))
		dir := fmt.Sprintf("%x", hashDir[:8])
		dir = path.Join(".", "tmpdir", dir)

		wg.Add(1)
		go func(dir string, enumCount *uint64) {
			defer wg.Done()
			if err := prepareWorkSpace(pkgname, dir); err != nil {
				fmt.Printf("prepareWorkSpace failed: %s", err)
			}
			fmt.Println("prepare")
			if err := doVet(dir, pkgname, enumCount); err != nil {
				fmt.Printf("doVet failed: %s", err)
			}
		}(dir, &enumCount)
	}
	wg.Wait()

	fmt.Printf("enum count: %d\n", enumCount)
}

func prepareWorkSpace(pkgName, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// create go.mod
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		cmd := exec.Command("go", "mod", "init", "a")
		cmd.Dir = path.Join(".", dir)
		if err := cmd.Run(); err != nil {
			return errors.New(fmt.Sprintf("%s go mod init failed: %s", pkgName, err))
		}
	}
	return nil
}

// go mod init a
// go get pkgname/...
// go vet pkgname/...
func doVet(dir string, pkgname string, enumCount *uint64) error {
	defer func() {
		// remove dir
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("remove dir %s failed: %s", dir, err)
		}
	}()
	arg := path.Join(pkgname, "...")
	cmd := exec.Command("go", "get", arg)
	cmd.Dir = path.Join(".", dir)
	if err := cmd.Run(); err != nil {
		fmt.Printf("go get %s failed: %s", pkgname, err)
		return err
	}
	cmd = exec.Command("go", "vet", "-vettool=/Users/sugiurahajime/go/bin/enumResearch", arg)
	cmd.Dir = path.Join(".", dir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		print("vet output: ", string(out))
		atomic.AddUint64(enumCount, 1)
	}
	return nil
}
