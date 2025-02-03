package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/jmooli/pokedex/internal/pokecache"
	"io"
	"net/http"
	"os"
	"strings"
)

var commands map[string]cliCommand

type MapData struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ApiResponse struct {
	Count    int       `json:"count"`
	Next     string    `json:"next"`
	Previous string    `json:"previous"`
	Results  []MapData `json:"results"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(*Config) error
}

type Config struct {
	Next     *string
	Previous *string
}

func commandExit(c *Config) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(c *Config) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("")
	for _, v := range commands {
		fmt.Printf("%s: %s \n", v.name, v.description)
	}
	return nil
}

func commandMap(c *Config) error {
	apiResp, err := makeApiGetRequest(*c.Next)
	if err != nil {
		return err
	}
	printOutAreas(apiResp)
	c.Next = &apiResp.Next
	c.Previous = &apiResp.Previous
	return nil
}
func commandMapb(c *Config) error {
	apiResp, err := makeApiGetRequest(*c.Previous)
	if err != nil {
		return err
	}
	printOutAreas(apiResp)
	c.Next = &apiResp.Next
	c.Previous = &apiResp.Previous
	return nil
}
func makeApiGetRequest(url string) (ApiResponse, error) {
	var apiResp ApiResponse

	res, err := http.Get(url)
	if err != nil {
		return apiResp, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return apiResp, err
	}

	err = json.Unmarshal(data, &apiResp)
	if err != nil {
		return apiResp, err
	}
	return apiResp, nil
}

func printOutAreas(apiResp ApiResponse) {
	for _, v := range apiResp.Results {
		fmt.Printf("%s\n", v.Name)
	}
}

func init() {

	commands = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"map": {
			name:        "map",
			description: "Lists following 20 Maps",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "lists previous 20 Maps",
			callback:    commandMapb,
		},
	}
}

func main() {
	config := Config{}

	// init configuration with default values
	initConfig(&config)

	cache := Ca
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex >")
		scanner.Scan()
		input := scanner.Text()
		cleanedInput := cleanInput(input)

		if len(cleanedInput) == 0 {
			continue
		}

		command, exists := commands[cleanedInput[0]]
		if exists {
			err := command.callback(&config)
			if err != nil {
				fmt.Printf("Command error: %v\n", err)
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}

// should remove this, using the api to set next and previous @ command map and mapb
func initConfig(c *Config) {
	baseUrl := "https://pokeapi.co/api/v2/location-area/"
	fullUrl := baseUrl + "?offset=0&limit=20"
	c.Next = &fullUrl
}

func cleanInput(text string) []string {
	lower := strings.ToLower(text)
	splitString := strings.Fields(lower)
	return splitString
}
