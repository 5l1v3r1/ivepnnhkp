package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	pb "gopkg.in/cheggaaa/pb.v1"
)

func main() {
	directory := flag.String("directory", "", "")
	list := flag.String("list", "", "")
	template := flag.String("template", "", "")
	workers := flag.Int("workers", 0, "")
	flag.Parse()
	if *directory == "" {
		log.Println("Invalid Directory")
		return
	}
	if *list == "" {
		log.Println("Invalid List")
		return
	}
	if *template == "" {
		log.Println("Invalid Template")
		return
	}
	if *workers == 0 {
		log.Println("Invalid Workers")
		return
	}
	process(directory, list, template, workers)
}

func process(directory *string, list *string, template *string, workers *int) {
	waitGroup := &sync.WaitGroup{}
	files := getFiles(list)
	total := len(*files)
	contents := getContents(template)
	incoming := make(chan *string)
	outgoing := make(chan bool)
	waitGroup.Add(1)
	go producer(waitGroup, directory, files, incoming)
	for worker := 0; worker < *workers; worker++ {
		waitGroup.Add(1)
		go consumer(waitGroup, contents, incoming, outgoing)
	}
	waitGroup.Add(1)
	go reporter(waitGroup, &total, outgoing)
	waitGroup.Wait()
}

func getFiles(list *string) *[]string {
	files := &[]string{}
	resource, err := os.Open(*list)
	if err != nil {
		log.Fatal(err)
	}
	defer resource.Close()
	scanner := bufio.NewScanner(resource)
	for scanner.Scan() {
		file := scanner.Text()
		*files = append(*files, file)
	}
	return files
}

func getContents(template *string) *string {
	contentsBytes, err := ioutil.ReadFile(*template)
	if err != nil {
		contentsBytes = []byte("")
	}
	contentsString := string(contentsBytes)
	return &contentsString
}

func producer(waitGroup *sync.WaitGroup, directory *string, files *[]string, incoming chan *string) {
	defer waitGroup.Done()
	for _, file := range *files {
		path := getPath(directory, &file)
		incoming <- path
	}
	close(incoming)
}

func getPath(directory *string, file *string) *string {
	path := fmt.Sprintf("%s/%s", *directory, *file)
	return &path
}

func consumer(waitGroup *sync.WaitGroup, _contents *string, incoming chan *string, outgoing chan bool) {
	defer waitGroup.Done()
	for path := range incoming {
		_ = getContents(path)
		outgoing <- true
	}
}

func reporter(waitGroup *sync.WaitGroup, total *int, outgoing chan bool) {
	defer waitGroup.Done()
	bar := pb.StartNew(*total)
	count := 0
	for _ = range outgoing {
		bar.Increment()
		count++
		if count == *total {
			close(outgoing)
		}
	}
	bar.Finish()
}
