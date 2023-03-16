package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
)

func main() {
	// Make HTTP GET request
	response, err := http.Get("https://pkg.go.dev/go.uber.org/multierr?tab=importedby")
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	// grep ImportedBy-detailsIndent from response.Body
	// and write to standard output
	// read one line from response.Body
	scanner := bufio.NewScanner(response.Body)
	impoertedByCount := 0
	// int64 enumCounter
	var enumCount uint64 = 0
	wg := sync.WaitGroup{}

	for scanner.Scan() {
		// read one line from response.Body
		line := scanner.Text()
		// if line contains "ImportedBy-detailsIndent" then write to standard output
		if strings.Contains(line, "ImportedBy-detailsIndent") {
			impoertedByCount++
			// get url from line
			url := strings.Split(line, "href=\"")[1]
			url = strings.Split(url, "\"")[0]
			url, _ = strings.CutPrefix(url, "/")

			fmt.Println(url)

			prepareWorkSpace(url)
			fmt.Println("downloaded package")
			// "go vet  $(go list -f '{{.Dir}}' $(go list -f '{{join .Deps "\n"}}' a))"
			hashDir := sha256.Sum256([]byte(url))
			dir := fmt.Sprintf("%x", hashDir[:8])

			wg.Add(1)
			func(dir string, enumCount *uint64){
				defer wg.Done()
				doVet(dir, enumCount)
			}(dir, &enumCount)

			if impoertedByCount > 5 {
				break
			}
		}
	}
	wg.Wait()

	fmt.Printf("ImportedBy count: %d\n", impoertedByCount)
	fmt.Printf("enum count: %d\n", enumCount)
}

func prepareWorkSpace(pkgName string) {
	hashDir := sha256.Sum256([]byte(pkgName))
	// dir change to string
	dir := fmt.Sprintf("%x", hashDir[:8])
	// create dir
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal("mkdir failed: ", err)
	}
	// cd dir
	if err := os.Chdir(dir); err != nil {
		log.Fatal("cd failed: ", err)
	}
	// create go.mod
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		if err := exec.Command("go", "mod", "init", "a").Run(); err != nil {
			log.Fatal("go mod init failed: ", err)
		}
	}
	// make file and write url
	f, err := os.Create("tmp.go")
	if err != nil {
		log.Fatal("create tmp.go failed: ", err)
	}
	if _, err = f.WriteString("package a\nimport \"" + pkgName + "\""); err != nil {
		log.Fatal("write tmp.go failed: ", err)
	}

	if err = exec.Command("go", "mod", "tidy").Run(); err != nil {
		log.Fatal("go mod tidy failed: ", err)
	}
}

func doVet(dir string, enumCount *uint64) {
	// "go vet  $(go list -f '{{.Dir}}' $(go list -f '{{join .Deps "\n"}}' a))"
	cmd := exec.Command("zsh", "-c", `go list -f '{{.Dir}}' $(go list -f '{{join .Deps "\n"}}' a)`)
	pkglist, err := cmd.Output()
	if err != nil {
		log.Fatal("go list failed: ", err)
	}
	wg2 := sync.WaitGroup{}
	for _, dir := range strings.Split(string(pkglist), "\n") {
		wg2.Add(1)
		func (dir string) {
			cmd := exec.Command("zsh", "-c", `go vet -vettool=$(which enumResearch) `+dir)
			out, err := cmd.CombinedOutput()
			if err != nil {
				print("vet output: ", string(out))
				atomic.AddUint64(enumCount, 1)
			}
			wg2.Done()
		}(dir)
	}
	wg2.Wait()

	if err = os.Chdir(".."); err != nil {
		log.Fatal("cd failed: ", err)
	}
	// remove dir
	if err = os.RemoveAll(dir); err != nil {
		log.Fatal("remove failed: ", err)
	}
}
