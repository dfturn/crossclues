package codenames

import (
	"crypto/subtle"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jbowens/dictionary"
)

var closed chan struct{}

func init() {
	closed = make(chan struct{})
	close(closed)
}

type Server struct {
	Server http.Server
	Store  Store

	tpl         *template.Template
	gameIDWords []string

	mu           sync.Mutex
	games        map[string]*GameHandle
	defaultWords []string
	mux          *http.ServeMux

	statOpenRequests  int64 // atomic access
	statTotalRequests int64 // atomic access
}

type Store interface {
	Save(*Game) error
	Delete(*Game) error
	Checkpoint(io.Writer) error
}

type GameHandle struct {
	store Store

	mu        sync.Mutex
	updated   chan struct{} // closed when the game is updated
	replaced  chan struct{} // closed when the game has been replaced
	marshaled []byte
	g         *Game
}

func newHandle(g *Game, s Store) *GameHandle {
	gh := &GameHandle{
		store:    s,
		g:        g,
		updated:  make(chan struct{}),
		replaced: make(chan struct{}),
	}
	err := s.Save(g)
	if err != nil {
		log.Printf("Unable to write updated game %q to disk: %s\n", gh.g.ID, err)
	}
	return gh
}

func (gh *GameHandle) update(fn func(*Game) bool) {
	gh.mu.Lock()
	defer gh.mu.Unlock()
	ok := fn(gh.g)
	if !ok {
		// game wasn't updated
		return
	}

	gh.marshaled = nil
	ch := gh.updated
	gh.updated = make(chan struct{})

	// write the updated game to disk
	err := gh.store.Save(gh.g)
	if err != nil {
		log.Printf("Unable to write updated game %q to disk: %s\n", gh.g.ID, err)
	}

	close(ch)
}

func (gh *GameHandle) gameStateChanged(stateID *string) (updated <-chan struct{}, replaced <-chan struct{}) {
	if stateID == nil {
		return closed, nil
	}

	gh.mu.Lock()
	defer gh.mu.Unlock()
	if gh.g.StateID() != *stateID {
		return closed, nil
	}
	return gh.updated, gh.replaced
}

// MarshalJSON implements the encoding/json.Marshaler interface.
// It caches a marshalled value of the game object.
func (gh *GameHandle) MarshalJSON() ([]byte, error) {
	gh.mu.Lock()
	defer gh.mu.Unlock()

	var err error
	if gh.marshaled == nil {
		gh.marshaled, err = json.Marshal(struct {
			*Game
			StateID string `json:"state_id"`
		}{gh.g, gh.g.StateID()})
	}
	return gh.marshaled, err
}

func (s *Server) getGame(gameID string) *GameHandle {
	s.mu.Lock()
	defer s.mu.Unlock()

	gh, ok := s.games[gameID]
	if ok {
		return gh
	}
	gh = newHandle(newGame(gameID, randomState(s.defaultWords, DefaultBoardSize), GameOptions{}), s.Store)
	s.games[gameID] = gh
	return gh
}

// POST /game-state
func (s *Server) handleGameState(rw http.ResponseWriter, req *http.Request) {
	var body struct {
		GameID   string  `json:"game_id"`
		StateID  *string `json:"state_id"`
		PlayerID string  `json:"player_id"`
	}
	err := json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		http.Error(rw, "Error decoding request body", 400)
		return
	}

	gh := s.getGame(body.GameID)

	gh.update(func(g *Game) bool {
		err = g.Draw(body.PlayerID)
		return err == nil
	})
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}

	updated, replaced := gh.gameStateChanged(body.StateID)

	select {
	case <-req.Context().Done():
		return
	case <-time.After(15 * time.Second):
		writeGameForPlayer(rw, gh, body.PlayerID)
	case <-updated:
		writeGameForPlayer(rw, gh, body.PlayerID)
	case <-replaced:
		gh = s.getGame(body.GameID)
		writeGameForPlayer(rw, gh, body.PlayerID)
	}
}

// POST /guess
func (s *Server) handleGuess(rw http.ResponseWriter, req *http.Request) {
	var request struct {
		GameID   string `json:"game_id"`
		Index    int    `json:"index"`
		PlayerID string `json:"player_id"`
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		http.Error(rw, "Error decoding", 400)
		return
	}

	gh := s.getGame(request.GameID)

	var err error
	gh.update(func(g *Game) bool {
		err = g.Guess(request.Index, request.PlayerID)
		return err == nil
	})
	if err != nil {
		http.Error(rw, err.Error(), 400)
		return
	}
	writeGameForPlayer(rw, gh, request.PlayerID)
}

func (s *Server) handleNextGame(rw http.ResponseWriter, req *http.Request) {
	var request struct {
		GameID          string   `json:"game_id"`
		PlayerID        string   `json:"player_id"`
		WordSet         []string `json:"word_set"`
		CreateNew       bool     `json:"create_new"`
		TimerDurationMS int64    `json:"timer_duration_ms"`
		EnforceTimer    bool     `json:"enforce_timer"`
		HandSize        int      `json:"hand_size"`
		BoardSize       int      `json:"board_size"`
	}

	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(rw, "Error decoding", 400)
		return
	}
	wordSet := map[string]bool{}
	for _, w := range request.WordSet {
		wordSet[strings.TrimSpace(strings.ToUpper(w))] = true
	}
	if len(wordSet) > 0 && len(wordSet) < 25 {
		http.Error(rw, "Need at least 25 words", 400)
		return
	}
	if len(wordSet) > 10000 {
		http.Error(rw, "Too many words in the set.", 400)
		return
	}

	var gh *GameHandle
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		words := s.defaultWords
		if len(wordSet) > 0 {
			words = nil
			for w := range wordSet {
				words = append(words, w)
			}
			sort.Strings(words)
		}

		opts := GameOptions{
			TimerDurationMS: request.TimerDurationMS,
			EnforceTimer:    request.EnforceTimer,
			HandSize:        request.HandSize,
			BoardSize:       request.BoardSize,
		}

		var ok bool
		gh, ok = s.games[request.GameID]
		if !ok {
			// no game exists, create for the first time
			gh = newHandle(newGame(request.GameID, randomState(words, opts.BoardSize), opts), s.Store)
			s.games[request.GameID] = gh
		} else if request.CreateNew {
			replacedCh := gh.replaced

			previousGame := gh.g

			nextState := nextGameState(gh.g.GameState, gh.g.BoardSize)
			gh = newHandle(newGame(request.GameID, nextState, opts), s.Store)
			s.games[request.GameID] = gh

			// signal to waiting /game-state goroutines that the
			// old game was swapped out for a new game.
			close(replacedCh)

			// Delete the old game from the store. This isn't strictly
			// necessary, but it helps us reclaim disk space a little more
			// quickly.
			err := s.Store.Delete(previousGame)
			if err != nil {
				log.Printf("Unable to delete old game %q from disk: %s\n", previousGame.ID, err)
			}
		}
	}()
	writeGameForPlayer(rw, gh, request.PlayerID)
}

type statsResponse struct {
	GamesTotal          int   `json:"games_total"`
	GamesInProgress     int   `json:"games_in_progress"`
	GamesCreatedOneHour int   `json:"games_created_1h"`
	RequestsTotal       int64 `json:"requests_total_process_lifetime"`
	RequestsInFlight    int64 `json:"requests_in_flight"`
}

func (s *Server) handleStats(rw http.ResponseWriter, req *http.Request) {
	hourAgo := time.Now().Add(-time.Hour)

	s.mu.Lock()
	defer s.mu.Unlock()

	var inProgress, createdWithinAnHour int
	for _, gh := range s.games {
		gh.mu.Lock()
		if gh.g.Won == false && gh.g.anyRevealed() {
			inProgress++
		}
		if hourAgo.Before(gh.g.CreatedAt) {
			createdWithinAnHour++
		}
		gh.mu.Unlock()
	}
	writeJSON(rw, statsResponse{
		GamesTotal:          len(s.games),
		GamesInProgress:     inProgress,
		GamesCreatedOneHour: createdWithinAnHour,
		RequestsTotal:       atomic.LoadInt64(&s.statTotalRequests),
		RequestsInFlight:    atomic.LoadInt64(&s.statOpenRequests),
	})
}

func (s *Server) handleCheckpoint(rw http.ResponseWriter, req *http.Request) {
	err := s.Store.Checkpoint(rw)
	if err != nil {
		log.Printf("[ERROR] Write checkpoint %s\n", err)
	}
}

func (s *Server) cleanupOldGames() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, gh := range s.games {
		gh.mu.Lock()
		if gh.g.Won == false && gh.g.CreatedAt.Add(3*time.Hour).Before(time.Now()) {
			delete(s.games, id)
			log.Printf("Removed completed game %s\n", id)
		} else if gh.g.CreatedAt.Add(72 * time.Hour).Before(time.Now()) {
			delete(s.games, id)
			log.Printf("Removed expired game %s\n", id)
		}
		gh.mu.Unlock()
	}
}

func (s *Server) Start(games map[string]*Game) error {
	gameIDs, err := dictionary.Load("assets/game-id-words.txt")
	if err != nil {
		return err
	}
	d, err := dictionary.Load("assets/original.txt")
	if err != nil {
		return err
	}
	s.tpl, err = template.New("index").Parse(tpl)
	if err != nil {
		return err
	}

	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/stats", s.handleStats)
	s.mux.HandleFunc("/next-game", s.handleNextGame)
	s.mux.HandleFunc("/guess", s.handleGuess)
	s.mux.HandleFunc("/game-state", s.handleGameState)
	s.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend/dist"))))
	s.mux.HandleFunc("/", s.handleIndex)

	bootstrapPW := os.Getenv("BOOTSTRAPPW")
	// If no bootstrap PW is set, don't expose the checkpoint endpoint so we
	// don't default to open.
	if bootstrapPW != "" {
		log.Printf("/checkpoint endpoint enabled\n")
		s.mux.Handle("/checkpoint", basicAuth(
			http.HandlerFunc(s.handleCheckpoint),
			os.Getenv("BOOTSTRAPPW"),
			"admin"))
	}

	gameIDs = dictionary.Filter(gameIDs, func(s string) bool { return len(s) >= 3 })
	s.gameIDWords = gameIDs.Words()
	for i, w := range s.gameIDWords {
		s.gameIDWords[i] = strings.ToLower(w)
	}

	s.games = make(map[string]*GameHandle)
	s.defaultWords = d.Words()
	sort.Strings(s.defaultWords)
	s.Server.Handler = withPProfHandler(s)

	if s.Store == nil {
		s.Store = discardStore{}
	}

	if games != nil {
		for _, g := range games {
			s.games[g.ID] = newHandle(g, s.Store)
		}
	}

	go func() {
		for range time.Tick(10 * time.Minute) {
			s.cleanupOldGames()
		}
	}()

	return s.Server.ListenAndServe()
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	atomic.AddInt64(&s.statTotalRequests, 1)
	atomic.AddInt64(&s.statOpenRequests, 1)
	defer func() { atomic.AddInt64(&s.statOpenRequests, -1) }()

	s.mux.ServeHTTP(rw, req)
}

func withPProfHandler(next http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	pprofHandler := basicAuth(mux, os.Getenv("PPROFPW"), "admin")

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/debug/pprof") {
			pprofHandler.ServeHTTP(rw, req)
			return
		}
		next.ServeHTTP(rw, req)
	})
}

func basicAuth(handler http.Handler, password, realm string) http.Handler {
	p := []byte(password)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(pass), p) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			io.WriteString(w, "Unauthorized\n")
			return
		}
		handler.ServeHTTP(w, r)
	})
}

func writeGameForPlayer(rw http.ResponseWriter, gh *GameHandle, playerID string) {
	gameCopy := gh.g.ClientCopy(playerID)
	writeJSON(rw, gameCopy)
}

func writeJSON(rw http.ResponseWriter, resp interface{}) {
	j, err := json.Marshal(resp)
	if err != nil {
		http.Error(rw, "unable to marshal response: "+err.Error(), 500)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(j)
}
