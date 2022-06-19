package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Paper struct {
	URL      string
	Title    string
	Category string
	Filename string
}

type Card struct {
	Name    string `json: "name"`
	SpaceId string `json: "spaceId"`
}

func parseInput() Paper {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter PDF URL: ")
	url, _ := reader.ReadString('\n')

	fmt.Print("Enter paper title: ")
	title, _ := reader.ReadString('\n')

	fmt.Print("Enter category (optional): ")
	category, _ := reader.ReadString('\n')
	fmt.Print(category)
	if category == "\n" {
		category = "Miscellaneous\n"
	}

	paper := Paper{
		URL:      strings.TrimSuffix(url, "\n"),
		Title:    strings.TrimSuffix(title, "\n"),
		Category: strings.TrimSuffix(category, "\n"),
	}
	paper.Filename = strings.Join(strings.Split(strings.ToLower(strings.ReplaceAll(paper.Title, ",", "")), " "), "-") + ".pdf"

	return paper
}

func downloadPaper(paper Paper) error {
	// create output file
	out, err := os.Create(fmt.Sprintf("/Users/max/Documents/Papers/%s", paper.Filename))
	if err != nil {
		return err
	}

	// get data
	resp, err := http.Get(paper.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// copy the data to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func commitPaper(paper Paper) error {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("git add . && git commit -m 'adds paper %s' && git push", paper.Title))
	cmd.Dir = "/Users/max/Documents/Papers"
	return cmd.Run()
}

func createCard(paper Paper) Card {
	return Card{
		Name:    fmt.Sprintf("[%s](https://raw.githubusercontent.com/maxtaylordavies/papers/master/%s)", paper.Title, paper.Filename),
		SpaceId: "papers-4oKyeUTNlswo5j4hw1sQP",
	}
}

func addToKinopio(paper Paper) error {
	token := "b6393d27-9473-4910-b630-82b1baee20d6"

	// encode payload
	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(createCard(paper))

	// create POST request
	req, err := http.NewRequest("POST", "https://api.kinopio.club/card", buffer)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// dispatch request
	client := &http.Client{}
	_, err = client.Do(req)

	return err
}

func main() {
	// get required info from stdin and create a Paper object
	paper := parseInput()

	// download the pdf into Documents/Papers
	err := downloadPaper(paper)
	if err != nil {
		log.Fatal(err)
	}

	// commit change to git + push to github
	err = commitPaper(paper)
	if err != nil {
		log.Fatal(err)
	}

	// add paper to kinopio
	err = addToKinopio(paper)
	if err != nil {
		log.Fatal(err)
	}
}
