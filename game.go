package crossclues

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

const DefaultBoardSize = 4

// GameState encapsulates enough data to reconstruct
// a Game's state. It's used to recreate games after
// a process restart.
type GameState struct {
	Seed        int64          `json:"seed"`
	PermIndex   int            `json:"perm_index"`
	Revealed    []bool         `json:"revealed"`
	WordSet     []string       `json:"word_set"`
	DeckIndex   int            `json:"deck_index"`
	PlayerCards map[int]string `json:"player_cards"`
	Discards    map[int]string `json:"discards"`
}

func (gs GameState) anyRevealed() bool {
	var revealed bool
	for _, r := range gs.Revealed {
		revealed = revealed || r
	}
	return revealed
}

func randomState(words []string, boardSize int) GameState {
	return GameState{
		Seed:        rand.Int63(),
		PermIndex:   0,
		Revealed:    make([]bool, getTotalSpaces(boardSize)),
		WordSet:     words,
		DeckIndex:   0,
		PlayerCards: make(map[int]string),
		Discards:    make(map[int]string),
	}
}

func getWordsPerGame(boardSize int) int {
	return boardSize * 2
}

func getTotalSpaces(boardSize int) int {
	return boardSize * boardSize
}

// nextGameState returns a new GameState for the next game.
func nextGameState(state GameState, boardSize int) GameState {
	wordsPerGame := getWordsPerGame(boardSize)
	state.PermIndex = state.PermIndex + wordsPerGame
	if state.PermIndex+wordsPerGame >= len(state.WordSet) {
		state.Seed = rand.Int63()
		state.PermIndex = 0
	}
	state.Revealed = make([]bool, getTotalSpaces(boardSize))
	state.DeckIndex = 0
	state.PlayerCards = make(map[int]string)
	state.Discards = make(map[int]string)
	return state
}

type Game struct {
	GameState
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Words     []string  `json:"words"`
	Deck      []int     `json:"deck"`
	Score     int       `json:"score"`
	Won       bool      `json:"won"` // TODO: Update name to "gameOver"
	GameOptions
}

type GameOptions struct {
	TimerDurationMS int64 `json:"timer_duration_ms,omitempty"`
	EnforceTimer    bool  `json:"enforce_timer,omitempty"`
	HandSize        int   `json:"hand_size,omitempty"`
	BoardSize       int   `json:"board_size,omitempty"`
}

func (g *Game) ClientCopy(playerID string) Game {
	var newGame Game

	newGame.Seed = g.Seed
	newGame.PermIndex = g.PermIndex
	newGame.Revealed = append(newGame.Revealed, g.Revealed...)
	newGame.DeckIndex = g.DeckIndex
	newGame.PlayerCards = make(map[int]string)
	for key, val := range g.PlayerCards {
		if val != playerID {
			continue
		}
		newGame.PlayerCards[key] = val
	}

	newGame.Discards = make(map[int]string)
	for key, val := range g.Discards {
		if val != playerID {
			continue
		}
		newGame.Discards[key] = val
	}

	newGame.ID = g.ID
	newGame.CreatedAt = g.CreatedAt
	newGame.UpdatedAt = g.UpdatedAt
	newGame.Words = append(newGame.Words, g.Words...)
	newGame.Score = g.Score
	newGame.Won = g.Won

	newGame.TimerDurationMS = g.TimerDurationMS
	newGame.EnforceTimer = g.EnforceTimer
	newGame.HandSize = g.HandSize
	newGame.BoardSize = g.BoardSize
	return newGame
}

func (g *Game) StateID() string {
	return fmt.Sprintf("%019d", g.UpdatedAt.UnixNano())
}

// TODO: Test some of this stuff
func (g *Game) checkWinningCondition() {
	score := 0
	for _, r := range g.Revealed {
		if !r {
			continue
		}
		score++
	}
	g.Score = score

	playedCardCount := g.DeckIndex - len(g.PlayerCards)
	g.Won = playedCardCount == getTotalSpaces(g.BoardSize)
}

func (g *Game) Discard(playerID string, idx int) error {
	if g.Won {
		return nil
	}

	cardPlayerID := g.PlayerCards[idx]
	if cardPlayerID != playerID {
		return fmt.Errorf("index %d is not owned by player %s", idx, playerID)
	}

	g.UpdatedAt = time.Now()

	g.Discards[idx] = playerID
	delete(g.PlayerCards, idx)
	err := g.Draw(playerID)

	g.checkWinningCondition()
	return err
}

func (g *Game) Draw(playerID string) error {
	if g.Won {
		return nil
	}

	cardCount := 0
	for _, value := range g.PlayerCards {
		if value == playerID {
			cardCount++
		}
	}

	if cardCount >= g.HandSize {
		return nil
	}

	if g.DeckIndex >= len(g.Deck) {
		// It's fine to run out of cards
		return nil
	}

	g.UpdatedAt = time.Now()

	for ; cardCount < g.HandSize; cardCount++ {
		card := g.Deck[g.DeckIndex]
		g.DeckIndex++

		g.PlayerCards[card] = playerID
	}

	g.checkWinningCondition()
	return nil
}

func (g *Game) Guess(idx int, playerID string) error {
	if idx >= len(g.Revealed) || idx < 0 {
		return fmt.Errorf("index %d is invalid", idx)
	}
	if g.Revealed[idx] {
		return errors.New("cell has already been revealed")
	}
	cardPlayerID := g.PlayerCards[idx]
	if cardPlayerID != playerID {
		return fmt.Errorf("index %d is not owned by player %s", idx, playerID)
	}

	g.UpdatedAt = time.Now()
	g.Revealed[idx] = true

	// "play their card" and draw a new one
	delete(g.PlayerCards, idx)
	g.Draw(playerID)

	g.checkWinningCondition()

	return nil
}

func newGame(id string, state GameState, opts GameOptions) *Game {
	// consistent randomness across games with the same seed
	seedRnd := rand.New(rand.NewSource(state.Seed))
	// distinct randomness across games with same seed
	randRnd := rand.New(rand.NewSource(state.Seed * int64(state.PermIndex+1)))

	wordsPerGame := getWordsPerGame(opts.BoardSize)
	totalSpaces := getTotalSpaces(opts.BoardSize)

	game := &Game{
		ID:          id,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Words:       make([]string, 0, wordsPerGame),
		Deck:        make([]int, totalSpaces),
		Score:       0,
		Won:         false,
		GameState:   state,
		GameOptions: opts,
	}

	// Pick the next `wordsPerGame` words from the
	// randomly generated permutation
	perm := seedRnd.Perm(len(state.WordSet))
	permIndex := state.PermIndex
	for _, i := range perm[permIndex : permIndex+wordsPerGame] {
		w := state.WordSet[perm[i]]
		game.Words = append(game.Words, w)
	}

	// Pick a random permutation of the deck
	deck := make([]int, totalSpaces)
	for i := 0; i < totalSpaces; i++ {
		deck[i] = i
	}

	shuffleCount := randRnd.Intn(5) + 5
	for i := 0; i < shuffleCount; i++ {
		shuffle(randRnd, deck)
	}
	game.Deck = deck
	return game
}

func shuffle(rnd *rand.Rand, deck []int) {
	for i := range deck {
		j := rnd.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
}
