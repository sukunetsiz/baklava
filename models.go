package main

import "html/template"

// CellStyle holds style parameters for a cell.
type CellStyle struct {
	Rotation    float64
	Hue         float64
	Saturation  float64
	Lightness   float64
	HopDuration float64
	HopDelay    float64
}

// Cell represents a grid cell.
type Cell struct {
	Letter string
	Styles CellStyle
}

// PageData is passed to the HTML templates.
type PageData struct {
	Grid             [][]Cell
	GameLetter       string
	Size             int
	MinCoord         int
	MaxCoord         int
	Message          string
	Example          *CoordinateExample
	ShortInstruction string
}

// CoordinateExample is used to show an example coordinate for a letter.
type CoordinateExample struct {
	Letter      string
	X           int
	Y           int
	Explanation string
}

// ExampleLetter is a helper type used when selecting an example from the grid.
type ExampleLetter struct {
	Letter string
	X      int
	Y      int
}

const (
	GRID_SIZE            = 8
	MIN_COORDINATE       = 1
	MAX_COORDINATE       = 8
	RANDOM_LETTER_CHANCE = 30
)

// StyleRanges holds min and max values for each style attribute.
var StyleRanges = map[string][2]float64{
	"hue":         {0, 360},
	"saturation":  {70, 100},
	"lightness":   {60, 80},
	"hopDuration": {1.5, 2.5},
	"hopDelay":    {0, 2.0},
	"rotation":    {-10, 10},
}

// Global variable for templates.
var templates *template.Template

