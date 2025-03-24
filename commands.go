package main

import (
	"encoding/json"
	"fmt"
	"github.com/neccarus/pokedex/internal/pokecache"
	"github.com/neccarus/pokedex/internal/types"
	"io"
	"math/rand"
	"net/http"
	"os"
)

type cliCommand struct {
	name        string
	description string
	callback    func(conf *config, cache *pokecache.Cache) error
}

type cliArgCommand struct {
	cliCommand
	callback func(conf *config, cache *pokecache.Cache, args []string) error
}

var commands map[string]cliCommand
var argCommands map[string]cliArgCommand
var caughtPokemon map[string]types.Pokemon

func init() { // Need to use init() function to prevent initialization loop between commands and commandHelp
	commands = map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Display the locations in pokemon world",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Display the locations in pokemon world goes to previous map list",
			callback:    commandMapb,
		},
		"pokedex": {
			name:        "pokedex",
			description: "Display the Pokemon you have caught in your Pokedex",
			callback:    commandPokedex,
		},
	}
	argCommands = map[string]cliArgCommand{
		"explore": {
			cliCommand: cliCommand{
				name:        "explore",
				description: "Explore a given location in the Pokemon world",
			},
			callback: commandExplore,
		},
		"catch": {
			cliCommand: cliCommand{
				name:        "catch",
				description: "Attempt to catch a Pokemon and add it to your Pokedex",
			},
			callback: commandCatch,
		},
		"inspect": {
			cliCommand: cliCommand{
				name:        "inspect",
				description: "Inspect any pokemon in your Pokedex",
			},
			callback: commandInspect,
		},
	}
	caughtPokemon = map[string]types.Pokemon{}
}

type config struct {
	Next     string
	Previous string
}

type locationCall struct {
	Count    int
	Next     string
	Previous string
	Results  []result
}
type result struct {
	Name string
	URL  string
}

func commandExit(conf *config, cache *pokecache.Cache) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(conf *config, cache *pokecache.Cache) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("")
	for _, command := range commands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	for _, command := range argCommands {
		fmt.Printf("%s: %s\n", command.name, command.description)
	}
	return nil
}

func commandMap(conf *config, cache *pokecache.Cache) error {
	data, err := getAPIData(conf.Next, cache)
	if err != nil {
		return err
	}

	locationInfo := locationCall{}
	if err := json.Unmarshal(data, &locationInfo); err != nil {
		return err
	}

	conf.Next = locationInfo.Next
	conf.Previous = locationInfo.Previous

	for _, location := range locationInfo.Results {
		fmt.Println(location.Name)
	}
	return nil
}

func commandMapb(conf *config, cache *pokecache.Cache) error {
	if conf.Previous == "" {
		fmt.Println("you're on the first page")
		return nil
	}
	data, err := getAPIData(conf.Previous, cache)
	if err != nil {
		return err
	}

	locationInfo := locationCall{}
	if err := json.Unmarshal(data, &locationInfo); err != nil {
		return err
	}
	conf.Next = locationInfo.Next
	conf.Previous = locationInfo.Previous
	for _, location := range locationInfo.Results {
		fmt.Println(location.Name)
	}
	return nil
}

func commandExplore(conf *config, cache *pokecache.Cache, args []string) error {
	baseUrl := "https://pokeapi.co/api/v2/location-area/"
	name := args[0]
	exploreUrl := baseUrl + name
	data, err := getAPIData(exploreUrl, cache)
	if err != nil {
		return err
	}
	areaInfo := types.AreaInfo{}
	if err = json.Unmarshal(data, &areaInfo); err != nil {
		return err
	}
	fmt.Printf("Exploring %s...\nFound Pokemon:\n", name)
	for _, encounter := range areaInfo.PokemonEncounters {
		fmt.Println(encounter.Pokemon.Name)
	}

	return nil
}

func commandCatch(conf *config, cache *pokecache.Cache, args []string) error {
	baseUrl := "https://pokeapi.co/api/v2/pokemon/"
	name := args[0]
	catchUrl := baseUrl + name
	data, err := getAPIData(catchUrl, cache)
	if err != nil {
		return err
	}
	pokemonInfo := types.Pokemon{}
	if err = json.Unmarshal(data, &pokemonInfo); err != nil {
		return err
	}
	captureFactor := pokemonInfo.BaseExperience
	fmt.Printf("Throwing a Pokeball at %s...\n", name)
	if rand.Intn(captureFactor) >= captureFactor-captureFactor/3 {
		caughtPokemon[name] = pokemonInfo
		fmt.Printf("%s was caught!\n", name)
		fmt.Println("You may now inspect it with the inspect command.")
	} else {
		fmt.Printf("%s escaped!\n", name)
	}
	return nil
}

func commandInspect(conf *config, cache *pokecache.Cache, args []string) error {
	name := args[0]
	data, ok := caughtPokemon[name]
	if !ok {
		fmt.Printf("%s has not been caught, or is not a Pokemon", name)
		return nil
	}

	fmt.Printf("Name: %s\n", data.Name)
	fmt.Printf("Height: %d\n", data.Height)
	fmt.Printf("Weight: %d\n", data.Weight)
	fmt.Println("Stats:")
	for _, stat := range data.Stats {
		fmt.Printf("\t-%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, typ := range data.Types {
		fmt.Printf("\t- %s\n", typ.Type.Name)
	}

	return nil
}

func commandPokedex(cond *config, cache *pokecache.Cache) error {
	if len(caughtPokemon) == 0 {
		fmt.Println("No Pokemon were found")
		return nil
	}
	fmt.Println("Your Pokedex:")
	for _, pokemon := range caughtPokemon {
		fmt.Printf(" - %s\n", pokemon.Name)
	}
	return nil
}

func getAPIData(url string, cache *pokecache.Cache) ([]byte, error) {
	var data []byte
	var ok bool
	if data, ok = cache.Get(url); ok {
		fmt.Println("Cache hit")
	} else {
		res, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode > 200 {
			return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
		}
		data, err = io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
	}
	cache.Add(url, data)

	return data, nil
}
