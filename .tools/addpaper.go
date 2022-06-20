package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	kinopigo "github.com/maxtaylordavies/kinopigo"
)

const spaceID = "4oKyeUTNlswo5j4hw1sQP"

type Paper struct {
	URL      string
	Title    string
	Category string
	Filename string
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

// func getSpace() (Space, error) {
// 	var space Space

// 	// create GET request
// 	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.kinopio.club/space/%s", spaceID), nil)
// 	req.Header.Set("Authorization", token)

// 	// dispatch request
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return space, err
// 	}

// 	// attempt to decode the response body into a Space struct
// 	err = json.NewDecoder(resp.Body).Decode(&space)
// 	return space, err
// }

func createCard(paper Paper, space kinopigo.Space) kinopigo.Card {
	category := strings.ToLower(strings.ReplaceAll(paper.Category, " ", ""))
	var parent kinopigo.Card
	for _, c := range space.Cards {
		name := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(c.Name, "#", ""), " ", ""))
		if name == category {
			log.Println(c.Name)
			parent = c
			break
		}
	}

	return kinopigo.Card{
		Name:     fmt.Sprintf("[%s](https://raw.githubusercontent.com/maxtaylordavies/papers/master/%s)", paper.Title, paper.Filename),
		SpaceID:  spaceID,
		ParentID: parent.ID,
		X:        parent.X + 10,
		Y:        parent.Y + 10,
		Z:        parent.Z,
	}
}

func createConnection(parentID string, childID string, space kinopigo.Space) kinopigo.Connection {
	// determine what ConnectionTypeID we should use
	var ctid string
	for _, conn := range space.Connections {
		if conn.StartCardID == parentID {
			ctid = conn.ConnectionTypeID
			break
		}
	}

	// create an instance of Connection
	return kinopigo.Connection{
		SpaceID:          space.ID,
		ConnectionTypeID: ctid,
		StartCardID:      parentID,
		EndCardID:        childID,
	}
}

func addToKinopio(paper Paper) error {
	// create instance of Kinopigo client
	client, err := kinopigo.NewKinopigoClient()
	if err != nil {
		return err
	}

	space, err := client.GetSpace(spaceID)
	if err != nil {
		return err
	}

	// create card
	card := createCard(paper, space)
	card, err = client.CreateCard(card)
	if err != nil {
		return err
	}

	// create connection
	connection := createConnection(card.ParentID, card.ID, space)
	connection, err = client.CreateConnection(connection)
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
