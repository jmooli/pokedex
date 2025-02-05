package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jmooli/pokedex/internal/pokecache"
)

var commands map[string]cliCommand
var pokedex map[string]ApiPokemonResponse

type Result struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type cliCommand struct {
	name        string
	description string
	callback    func(c *Config, param string) error
}

type Config struct {
	Next     *string
	Previous *string
	Cache    *pokecache.Cache
}

func commandExit(c *Config, param string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(c *Config, param string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("")
	for _, v := range commands {
		fmt.Printf("%s: %s \n", v.name, v.description)
	}
	return nil
}

func commandMap(c *Config, param string) error {
	apiResp, err := makeApiGetRequest[ApiAreaResponse](*c.Next, c.Cache)
	if err != nil {
		return err
	}
	printOutAreaResponse(apiResp)
	c.Next = &apiResp.Next
	c.Previous = &apiResp.Previous
	return nil
}

func commandMapb(c *Config, param string) error {
	if c.Previous == nil || *c.Previous == "" {
		// no previous
		return nil
	}
	apiResp, err := makeApiGetRequest[ApiAreaResponse](*c.Previous, c.Cache)
	if err != nil {
		return err
	}
	printOutAreaResponse(apiResp)
	c.Next = &apiResp.Next
	c.Previous = &apiResp.Previous
	return nil
}

func commandExplore(c *Config, param string) error {
	baseUrl := "https://pokeapi.co/api/v2/location-area/"
	fullUlr := baseUrl + param
	fmt.Printf("Exploring %s...\n", param)
	apiResp, err := makeApiGetRequest[ApiEncounterResponse](fullUlr, c.Cache)
	if err != nil {
		return err
	}
	printOutPokemonResponce(apiResp)
	return nil
}

func commandCatch(c *Config, param string) error {
	baseUrl := "https://pokeapi.co/api/v2/pokemon/"
	fullUlr := baseUrl + param

	fmt.Printf("Throwing a Pokeball at %s...\n", param)

	apiresp, err := makeApiGetRequest[ApiPokemonResponse](fullUlr, c.Cache)

	if err != nil {
		return err
	}

	//calculation could be little bit more sophisticated
	if rand.Intn(300) > apiresp.BaseExperience {
		fmt.Printf("%s was caught!\n", param)

		// init if nil
		if pokedex == nil {
			pokedex = make(map[string]ApiPokemonResponse)
		}
		// store
		pokedex[param] = apiresp

	} else {
		fmt.Printf("%s escaped!\n", param)
	}

	return nil
}

func commandInspect(c *Config, param string) error {
	if pokemon, exists := pokedex[param]; exists {
		printPokemonDetails(pokemon)
	} else {
		fmt.Println("you have not caught that pokemon")
	}
	return nil
}

func printPokemonDetails(pokemon ApiPokemonResponse) {
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Println("Stats:")

	for _, stat := range pokemon.Stats {
		fmt.Printf("  - %s: %d\n", stat.Stat.Name, stat.BaseStat)
	}

	fmt.Println("Types:")
	for _, pokeType := range pokemon.Types {
		fmt.Printf("  - %s\n", pokeType.Type.Name)
	}
}

func makeApiGetRequest[T any](url string, c *pokecache.Cache) (T, error) {
	var result T

	data, found := c.Get(url)

	if !found {
		resp, err := http.Get(url)
		if err != nil {
			return result, err
		}
		defer resp.Body.Close()

		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return result, err
		}

		// store the new data
		c.Add(url, data)
	} else {
		//debug
		//fmt.Println("using cache")
	}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return result, err
	}
	return result, nil

}

func printOutAreaResponse(apiResp ApiAreaResponse) {
	for _, v := range apiResp.Results {
		fmt.Printf("%s\n", v.Name)
	}
}

func printOutPokemonResponce(apiResp ApiEncounterResponse) {
	if len(apiResp.PokemonEncounters) > 0 {
		fmt.Println("Found Pokemon:")
		for _, entry := range apiResp.PokemonEncounters {
			fmt.Printf(" - %s\n", entry.Pokemon.Name)
		}

	} else {
		fmt.Println("Didn't find any Pokemons:")
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
		"explore": {
			name:        "explore",
			description: "lists all pokemons in area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "attampt to catch a pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "Inspect Pokemon",
			callback:    commandInspect,
		},
	}
}

func main() {
	config := Config{}
	d := 30 * time.Second
	cache := pokecache.NewCache(d)
	initConfig(&config, cache)

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
		param := ""
		if len(cleanedInput) >= 2 {
			param = cleanedInput[1]
		}

		if exists {
			err := command.callback(&config, param)
			if err != nil {
				fmt.Printf("Command error: %v\n", err)
			}
		} else {
			fmt.Println("Unknown command")
		}
	}
}

// should move this init to main, using the api to set next and previous @ command map and mapb
func initConfig(c *Config, cache *pokecache.Cache) {
	baseUrl := "https://pokeapi.co/api/v2/location-area/"
	fullUrl := baseUrl + "?offset=0&limit=20"
	c.Next = &fullUrl
	c.Cache = cache
}

func cleanInput(text string) []string {
	lower := strings.ToLower(text)
	splitString := strings.Fields(lower)
	return splitString
}
