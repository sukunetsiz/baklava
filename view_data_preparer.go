package main

import (
	"fmt"
	"math/rand"

	"github.com/gorilla/sessions"
)

// prepareViewData creates a new captcha grid and updates the session with the captcha answer.
func prepareViewData(session *sessions.Session) PageData {
	grid, gameLetter, answer := generateCaptcha()
	session.Values["captcha_answer"] = answer
	session.Values["game_letter"] = gameLetter

	exampleLetter := getRandomExampleLetter(grid, gameLetter)
	data := PageData{
		Grid:             grid,
		GameLetter:       gameLetter,
		Size:             GRID_SIZE,
		MinCoord:         MIN_COORDINATE,
		MaxCoord:         MAX_COORDINATE,
		ShortInstruction: fmt.Sprintf("Find the appropriate box for the letter '%s'", gameLetter),
	}

	if exampleLetter != nil {
		data.Example = &CoordinateExample{
			Letter:      exampleLetter.Letter,
			X:           exampleLetter.X,
			Y:           exampleLetter.Y,
			Explanation: fmt.Sprintf("It is at coordinates %d-%d. This represents column %d and row %d", exampleLetter.X, exampleLetter.Y, exampleLetter.X, exampleLetter.Y),
		}
	}

	return data
}

// getRandomExampleLetter selects an example letter from the grid that is not the game letter.
func getRandomExampleLetter(grid [][]Cell, gameLetter string) *ExampleLetter {
	var examples []ExampleLetter
	for y := 0; y < GRID_SIZE; y++ {
		for x := 0; x < GRID_SIZE; x++ {
			if grid[y][x].Letter != "" && grid[y][x].Letter != gameLetter {
				examples = append(examples, ExampleLetter{
					Letter: grid[y][x].Letter,
					X:      x + 1,
					Y:      y + 1,
				})
			}
		}
	}
	if len(examples) == 0 {
		return nil
	}
	example := examples[rand.Intn(len(examples))]
	return &example
}

