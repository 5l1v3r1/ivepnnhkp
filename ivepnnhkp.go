package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sync"

	pb "gopkg.in/cheggaaa/pb.v1"
)

func main() {
	directory := flag.String("directory", "", "")
	list := flag.String("list", "", "")
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
	if *workers == 0 {
		log.Println("Invalid Workers")
		return
	}
	process(directory, list, workers)
}

func process(directory *string, list *string, workers *int) {
	waitGroup := &sync.WaitGroup{}
	files := getFiles(list)
	total := len(*files)
	regularExpression := regexp.MustCompile(`<\?php\s*\$ivepnnhkp.*?\?>`)
	incoming := make(chan *string)
	outgoing := make(chan bool)
	waitGroup.Add(1)
	go producer(waitGroup, directory, files, incoming)
	for worker := 0; worker < *workers; worker++ {
		waitGroup.Add(1)
		go consumer(waitGroup, regularExpression, incoming, outgoing)
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

func consumer(waitGroup *sync.WaitGroup, regularExpression *regexp.Regexp, incoming chan *string, outgoing chan bool) {
	defer waitGroup.Done()
	for path := range incoming {
		contents := getContents(path)
		if regularExpression.MatchString(*contents) {
			*contents = regularExpression.ReplaceAllString(*contents, "")
			setContents(path, contents)
		} else {
		}
		outgoing <- true
	}
}

func getContents(path *string) *string {
	contentsBytes, err := ioutil.ReadFile(*path)
	if err != nil {
		contentsBytes = []byte("")
	}
	contentsString := string(contentsBytes)
	return &contentsString
}

func setContents(path *string, contentsString *string) {
	contentsBytes := []byte(*contentsString)
	err := ioutil.WriteFile(*path, contentsBytes, 0644)
	if err != nil {
		log.Fatal(err)
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
