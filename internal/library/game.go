package library

type Source string

const (
	SourceGames  Source = "games"
	SourceArcade Source = "arcade"
)

type Game struct {
	ID           string
	Name         string
	Source       Source
	Year         string
	Manufacturer string
	Category     string
	RawCategory  string
	NonArcade    bool
	Vertical     bool
	ROM          string
	Command      string
	Args         []string
	WorkingDir   string
	CloneOf      string
}
