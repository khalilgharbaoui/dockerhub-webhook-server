package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"
	"time"
)

type Queue struct {
	items []*int
	size  int
	mutex *sync.Mutex
}

func (q *Queue) Push(item *int) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, item)
	q.size++
}

func (q *Queue) Shift() *int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	if q.size == 0 {
		return nil
	}

	top := q.items[0]
	q.items[0] = nil
	q.items = q.items[1:]

	q.size--

	return top
}

func (q *Queue) Size() int {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	return q.size
}

var q = &Queue{
	items: make([]*int, 0),
	size:  0,
	mutex: &sync.Mutex{},
}

var jobs = 0

func handler(w http.ResponseWriter, r *http.Request) {
	// Remove preceding slash from the URL and compare the rest against
	// AUTH_TOKEN value, if WEBHOOK_AUTH_TOKEN variable was not set Getenv
	// returns an empty string
	if r.URL.Path[1:] == os.Getenv("WEBHOOK_AUTH_TOKEN") {
		w.WriteHeader(http.StatusOK)

		// backlog size for the queue is 0; that means only one job can be
		// waiting in the queue
		if q.Size() == 0 {
			jobs++
			q.Push(&jobs)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func execAndStreamOutput(cmdName string, cmdArgs []string) {
	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("Error creating StdoutPipe for Cmd", err)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		log.Fatal("Error starting Cmd", err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal("Error waiting for Cmd", err)
	}
}

func setup() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	log.Println(dir + " <we are here")

	return path.Join(dir, "setup-helper.sh")
}

func processQueue() {
	for {
		if q.Size() == 0 {
			continue
		}

		job := *q.Shift()
		log.Printf("Processing job #%d\n", job)

		execAndStreamOutput("/bin/sh", []string{
			setup(),
		})

		log.Printf("Finished Processing job #%d\n", job)

		time.Sleep(3 * time.Second)
	}
}

func envError(name string) {
	log.Fatal(name + " environment variable must be set.")
}

func main() {
	// check environment variables
	requiredEnvVars := []string{
		"WEBHOOK_AUTH_TOKEN",
	}
	for _, item := range requiredEnvVars {
		if os.Getenv(item) == "" {
			envError(item)
		}
	}

	go processQueue()

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8008", nil)
	log.Printf("Listening on port: 8080......")
}
