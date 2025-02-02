package main

import (
    "encoding/json"
    "fmt"
    "html/template"
    "log"
    "math/rand"
    "net/http"
    "regexp"
    "strings"
    "time"
)

// Grid configuration constants
const (
    GRID_SIZE            = 8    // Size of the puzzle grid (8x8)
    MIN_COORDINATE       = 1    // Minimum coordinate value for user input
    MAX_COORDINATE       = 8    // Maximum coordinate value for user input
    RANDOM_LETTER_CHANCE = 30   // Percentage chance of random letter placement
)

// StyleRanges defines the visual appearance ranges for letters
// Each style parameter has a min and max value for randomization
var StyleRanges = map[string][2]float64{
    "hue":         {0, 360},    // Color hue range (0-360 degrees)
    "saturation":  {70, 100},   // Color saturation range (%)
    "lightness":   {60, 80},    // Color lightness range (%)
    "hopDuration": {1.5, 2.5},  // Animation duration range (seconds)
    "hopDelay":    {0, 2.0},    // Animation delay range (seconds)
    "rotation":    {-10, 10},   // Letter rotation range (degrees)
}

// CellStyle contains the visual styling information for each grid cell
type CellStyle struct {
    Rotation    float64
    Hue         float64
    Saturation  float64
    Lightness   float64
    HopDuration float64
    HopDelay    float64
}

// Cell represents a single grid position
type Cell struct {
    Letter string
    Styles CellStyle
}

// PageData contains all information needed for the template rendering
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

// CoordinateExample represents an example coordinate shown to users
type CoordinateExample struct {
    Letter      string
    X           int
    Y           int
    Explanation string
}

// Global variables
var (
    sessions  = make(map[string]string)  // Session storage (string-only)
    templates *template.Template          // Compiled templates
)

// init performs application initialization
func init() {
    // Initialize random seed
    rand.Seed(time.Now().UnixNano())
    
    // Define template functions
    funcMap := template.FuncMap{
        "iterate": func(start, end int) []int {
            var result []int
            for i := start; i < end; i++ {
                result = append(result, i)
            }
            return result
        },
        "add": func(a, b int) int {
            return a + b
        },
    }
    
    // Parse templates
    templates = template.Must(template.New("captcha.html").Funcs(funcMap).ParseFiles("templates/captcha.html"))
}

// main starts the HTTP server and sets up routes
func main() {
    // Set up static file serving
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    
    // Set up route handlers
    http.HandleFunc("/", showCaptcha)
    http.HandleFunc("/verify", verifyCaptcha)

    // Start the server
    log.Fatal(http.ListenAndServe(":8080", nil))
}

// showCaptcha handles the display of the CAPTCHA puzzle
func showCaptcha(w http.ResponseWriter, r *http.Request) {
    var data PageData
    
    // Check for existing CAPTCHA data in session
    if viewDataJSON, exists := sessions["view_data"]; exists {
        if err := json.Unmarshal([]byte(viewDataJSON), &data); err != nil {
            // Generate new data if unmarshaling fails
            data = prepareViewData()
            if jsonData, err := json.Marshal(data); err == nil {
                sessions["view_data"] = string(jsonData)
            }
        }
    } else {
        // Generate new data if none exists
        data = prepareViewData()
        if jsonData, err := json.Marshal(data); err == nil {
            sessions["view_data"] = string(jsonData)
        }
    }
    
    // Handle any message parameters
    if msg := r.URL.Query().Get("message"); msg != "" {
        data.Message = msg
    }
    
    // Render the template
    if err := templates.ExecuteTemplate(w, "captcha.html", data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}

// verifyCaptcha handles the validation of user answers
func verifyCaptcha(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
        http.Error(w, "Error parsing form", http.StatusBadRequest)
        return
    }

    answer := r.FormValue("captcha_answer")
    
    // Validate answer format
    if !regexp.MustCompile(`^[1-8]-[1-8]$`).MatchString(answer) {
        http.Redirect(w, r, "/?message=Please+enter+coordinates+in+correct+format", http.StatusSeeOther)
        return
    }

    // Compare answer with stored correct answer
    correctAnswer := sessions["captcha_answer"]
    formattedAnswer := formatCoordinateAnswer(answer)
    
    if formattedAnswer == correctAnswer {
        // Set solved status and clean up session
        sessions["captcha_solved"] = "true"
        delete(sessions, "captcha_answer")
        delete(sessions, "game_letter")
        delete(sessions, "view_data")
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }

    http.Redirect(w, r, "/?message=Incorrect+answer.+Please+try+again", http.StatusSeeOther)
}

// prepareViewData generates a new CAPTCHA puzzle and associated data
func prepareViewData() PageData {
    grid, gameLetter, answer := generateCaptcha()
    
    // Store answer and game letter in session
    sessions["captcha_answer"] = answer
    sessions["game_letter"] = gameLetter

    // Generate example for users
    exampleLetter := getRandomExampleLetter(grid, gameLetter)

    // Prepare page data
    data := PageData{
        Grid:             grid,
        GameLetter:       gameLetter,
        Size:             GRID_SIZE,
        MinCoord:         MIN_COORDINATE,
        MaxCoord:         MAX_COORDINATE,
        ShortInstruction: fmt.Sprintf("Find the appropriate box for the letter '%s'", gameLetter),
    }

    // Add example if available
    if exampleLetter != nil {
        data.Example = &CoordinateExample{
            Letter: exampleLetter.Letter,
            X:      exampleLetter.X,
            Y:      exampleLetter.Y,
            Explanation: fmt.Sprintf(
                "It is at coordinates %d-%d. This represents column %d and row %d. You can understand this by looking at the horizontal and vertical axes of the glowing letter.",
                exampleLetter.X, exampleLetter.Y, exampleLetter.X, exampleLetter.Y,
            ),
        }
    }

    return data
}

// generateCaptcha creates a new CAPTCHA puzzle grid
func generateCaptcha() ([][]Cell, string, string) {
    // Initialize empty grid
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

    // Generate random center position
    centerX := rand.Intn(6) + 1
    centerY := rand.Intn(6) + 1

    // Choose random game letter
    letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    gameLetter := string(letters[rand.Intn(len(letters))])

    // Define edges around center
    edges := [][2]int{
        {centerY - 1, centerX}, // Top
        {centerY + 1, centerX}, // Bottom
        {centerY, centerX - 1}, // Left
        {centerY, centerX + 1}, // Right
    }

    // Place edge letters
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

    // Handle center position
    if rand.Intn(2) == 0 {
        grid[centerY][centerX] = Cell{
            Letter: "",
            Styles: generateLetterStyles(),
        }
    } else {
        // Place random different letter in center
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

    // Create missing edge for puzzle
    missingEdgeIdx := rand.Intn(len(validEdges))
    missingEdge := validEdges[missingEdgeIdx]
    grid[missingEdge[0]][missingEdge[1]] = Cell{
        Letter: "",
        Styles: generateLetterStyles(),
    }
    
    // Calculate answer coordinates (add 1 to convert from 0-based to 1-based)
    answer := fmt.Sprintf("%d-%d", missingEdge[1]+1, missingEdge[0]+1)

    // Add random letters to grid
    addRandomLetters(grid, letters, centerX, centerY, edges)

    return grid, gameLetter, answer
}

// generateLetterStyles creates random visual styles for a letter
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

// generateRandomInRange returns a random float64 between min and max
func generateRandomInRange(min, max float64) float64 {
    return min + rand.Float64()*(max-min)
}

// isValidCoordinate checks if coordinates are within grid bounds
func isValidCoordinate(y, x int) bool {
    return y >= 0 && y < GRID_SIZE && x >= 0 && x < GRID_SIZE
}

// addRandomLetters fills empty grid spaces with random letters
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

// isEdgePosition checks if coordinates match any edge position
func isEdgePosition(y, x int, edges [][2]int) bool {
    for _, edge := range edges {
        if edge[0] == y && edge[1] == x {
            return true
        }
    }
    return false
}

// getRandomExampleLetter finds a random letter for the example
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

// ExampleLetter represents a letter used as an example
type ExampleLetter struct {
    Letter string
    X      int
    Y      int
}

// formatCoordinateAnswer standardizes the answer format
func formatCoordinateAnswer(answer string) string {
    return strings.ReplaceAll(answer, " ", "")
}
