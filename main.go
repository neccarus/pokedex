package main

import (
	"bufio"
	"fmt"
	"github.com/neccarus/pokedex/internal/pokecache"
	"os"
	"slices"
	"strings"
	"time"
)

func cleanInput(text string) []string {
	splitString := strings.Split(strings.Trim(text, " "), " ")
	var removeCache []int
	for index := range splitString {
		if len(splitString[index]) == 0 {
			removeCache = append([]int{index}, removeCache...)
			continue
		}
		splitString[index] = strings.ToLower(splitString[index])
	}
	for _, index := range removeCache {
		splitString = slices.Delete(splitString, index, index+1)
	}
	return splitString
}

func main() {

	scanner := bufio.NewScanner(os.Stdin)
	conf := &config{
		Next:     "https://pokeapi.co/api/v2/location-area",
		Previous: "",
	}
	cache := pokecache.NewCache(60 * time.Second)

	for {
		fmt.Print("Pokedex > ")
		if scanner.Scan() {
			text := scanner.Text()
			inputCommands := cleanInput(text)
			command := inputCommands[0]
			args := inputCommands[1:]
			//for _, command := range inputCommands {
			if _, ok := commands[command]; ok {
				if err := commands[command].callback(conf, cache); err != nil {
					fmt.Println(err)
				}
			} else if _, ok := argCommands[command]; ok {
				if err := argCommands[command].callback(conf, cache, args); err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Println("Unknown command")
			}
			//}
		}
	}
}
