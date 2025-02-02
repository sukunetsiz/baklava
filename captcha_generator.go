package main

import (
    "fmt"
    "math/rand"
)

func generateCaptcha() ([][]Cell, string, string) {
    grid := make([][]Cell, GRID_SIZE)
    for y := range grid {
        grid[y] = make([]Cell, GRID_SIZE)
        for x := range grid[y] {
            grid[y][x] = Cell{
                Letter: "",
                Styles: generateLetterStyles(),
            }
        }
    }

    centerX := rand.Intn(6) + 1
    centerY := rand.Intn(6) + 1
    letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    gameLetter := string(letters[rand.Intn(len(letters))])

    edges := [][2]int{
        {centerY - 1, centerX}, {centerY + 1, centerX},
        {centerY, centerX - 1}, {centerY, centerX + 1},
    }

    validEdges := make([][2]int, 0)
    for _, edge := range edges {
        if isValidCoordinate(edge[0], edge[1]) {
            grid[edge[0]][edge[1]] = Cell{
                Letter: gameLetter,
                Styles: generateLetterStyles(),
            }
            validEdges = append(validEdges, edge)
        }
    }

    if rand.Intn(2) == 0 {
        grid[centerY][centerX] = Cell{
            Letter: "",
            Styles: generateLetterStyles(),
        }
    } else {
        var centerLetter string
        for {
            centerLetter = string(letters[rand.Intn(len(letters))])
            if centerLetter != gameLetter {
                break
            }
        }
        grid[centerY][centerX] = Cell{
            Letter: centerLetter,
            Styles: generateLetterStyles(),
        }
    }

    missingEdgeIdx := rand.Intn(len(validEdges))
    missingEdge := validEdges[missingEdgeIdx]
    grid[missingEdge[0]][missingEdge[1]] = Cell{
        Letter: "",
        Styles: generateLetterStyles(),
    }

    answer := fmt.Sprintf("%d-%d", missingEdge[1]+1, missingEdge[0]+1)
    addRandomLetters(grid, letters, centerX, centerY, edges)

    return grid, gameLetter, answer
}

func generateLetterStyles() CellStyle {
    return CellStyle{
        Rotation:    generateRandomInRange(StyleRanges["rotation"][0], StyleRanges["rotation"][1]),
        Hue:         generateRandomInRange(StyleRanges["hue"][0], StyleRanges["hue"][1]),
        Saturation:  generateRandomInRange(StyleRanges["saturation"][0], StyleRanges["saturation"][1]),
        Lightness:   generateRandomInRange(StyleRanges["lightness"][0], StyleRanges["lightness"][1]),
        HopDuration: generateRandomInRange(StyleRanges["hopDuration"][0], StyleRanges["hopDuration"][1]),
        HopDelay:    generateRandomInRange(StyleRanges["hopDelay"][0], StyleRanges["hopDelay"][1]),
    }
}

func generateRandomInRange(min, max float64) float64 {
    return min + rand.Float64()*(max-min)
}

func addRandomLetters(grid [][]Cell, letters string, centerX, centerY int, edges [][2]int) {
    for y := 0; y < GRID_SIZE; y++ {
        for x := 0; x < GRID_SIZE; x++ {
            if (y == centerY && x == centerX) || isEdgePosition(y, x, edges) {
                continue
            }
            if grid[y][x].Letter == "" && rand.Intn(100) < RANDOM_LETTER_CHANCE {
                grid[y][x] = Cell{
                    Letter: string(letters[rand.Intn(len(letters))]),
                    Styles: generateLetterStyles(),
                }
            }
        }
    }
}

func isValidCoordinate(y, x int) bool {
    return y >= 0 && y < GRID_SIZE && x >= 0 && x < GRID_SIZE
}

func isEdgePosition(y, x int, edges [][2]int) bool {
    for _, edge := range edges {
        if edge[0] == y && edge[1] == x {
            return true
        }
    }
    return false
}
