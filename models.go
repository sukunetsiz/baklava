package main

import "html/template"

type CellStyle struct {
    Rotation    float64
    Hue         float64
    Saturation  float64
    Lightness   float64
    HopDuration float64
    HopDelay    float64
}

type Cell struct {
    Letter string
    Styles CellStyle
}

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

type CoordinateExample struct {
    Letter      string
    X           int
    Y           int
    Explanation string
}

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

var StyleRanges = map[string][2]float64{
    "hue":         {0, 360},
    "saturation":  {70, 100},
    "lightness":   {60, 80},
    "hopDuration": {1.5, 2.5},
    "hopDelay":    {0, 2.0},
    "rotation":    {-10, 10},
}

var (
    sessions  = make(map[string]string)
    templates *template.Template
)
