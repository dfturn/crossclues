package crossclues

import (
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
)

const tpl = `
<!DOCTYPE html>
<html>
    <head>
        <title>Cross Clues - Play Online</title>
        <script src="/static/app.js?v=0.02" type="text/javascript"></script>
        <link href="https://fonts.googleapis.com/css?family=Roboto" rel="stylesheet">
        <link rel="stylesheet" type="text/css" href="/static/game.css" />
        <link rel="stylesheet" type="text/css" href="/static/lobby.css" />
        <link rel="shortcut icon" type="image/png" id="favicon" href="data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAASABIAAD/2wBDABwcHBwcHDAcHDBEMDAwRFxEREREXHRcXFxcXHSMdHR0dHR0jIyMjIyMjIyoqKioqKjExMTExNzc3Nzc3Nzc3Nz/2wBDASIkJDg0OGA0NGDmnICc5ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubm5ubm5ub/wAARCACwATIDASIAAhEBAxEB/8QAGgAAAgMBAQAAAAAAAAAAAAAAAAQBAgMFBv/EADMQAAIBAgQEBAQHAAMBAAAAAAABAgMRBCExURITQXEFMmGBFCKRoSMzQlKxwdEVYnLw/8QAGAEBAQEBAQAAAAAAAAAAAAAAAAECAwT/xAAgEQEBAAIDAAMBAQEAAAAAAAAAAQIREiExA0FREyJh/9oADAMBAAIRAxEAPwD0M6sKfmZi8VBaJswxbtNdhO7uR0xxlm3S+KjsaKq30OUm8h642mUkb817E830MQJtltzfQOb6GIDY25q2DmrYxIuhsMc1BzUL3W4cS3G6hjmxDmxF009AeSuxuqY5sSObHYW447hxx3JyNGeatg5q2FeOO5aPzZocjRjnLYvGaloKtW1L0vMXYZAANIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA5+M8yfoJdR7G9OzEepK74eJWiHRFaD3T2Izm0WgvJviZutEYS8zMZsRanK0u5vVjePYVHIviimX479JlCIFpx4ZNA1ZL1O7zaVAAjJO6JbprDHd0tTbjJMblwyTVxNNXzJUr5GLlt6Z8evtFmFmTGaa7ETZz4tzEWNaT4XnoZppRuibp2bLMdJxhiUotahCajK7F+JImT2NHGeHedD1I58PUQuwuxteEPfEQ2ZHxEdmJfMFpbMbOEOfErYPiVsKcM9mHBPYbOOJr4l7EfEvZC/LqbE8qY2axMLE7oPiVsL8mYcmW6G01iY+KWxHxXoY8h7hyPUbNYtvi11iNRkpJSWjObKi4q6dxjCzunDbMsqZSa3DYABXMAAAAAACWM8qOf1Oli18i7nMJXbDxZDkdF2E1/Y3DyrsRM2sdDGfmZtHQzmrzRjJzhDEYunh3wO8pbLoRDxNUZ8vEU3H7nFqyc8VKT6z/ALOp4lS4qSqrWDs+zNSSUdabjOMasLSTB1Plskk0cbwqs25YWTyavHuTV8RhCTjTi5NdXkdXOy76dFszpu1Ro51PxHiko1I2T6o1nVccdGHRR/nMl8TCWXs+8pE6T75hPS5EtFL2Ob2ULKUolnnErLJqRZdUE+1YdUD8l9iqymaWzcdwVR5ovHOJmtMy9PqVPszRtnubitJ2n3GiMZ+ggkhdQyAAkCspxi0pNK+lyTneKK9GEtnY48K1Wn5JNGtI9UQcGHiVePmtI6eFxccTdW4WiaU2AAQQK0ny61nvYaFaytJS3DeH46YFKcuOCkXNuYAAAAAAFsUvw/c5Wx18Sr0mclkrrh4ByHlQmuo3T8qIufjWOhSeUky8epSr0M5OUeTrrgxE1tJ/yeklFVaTg9JROJ4lDhxPF0mk/wCjsYaXHQhL0Gd8pGfhmCVNc+p5+i2ObXw8f+QlRk7KTv8AXM9FSlZ23OL4vFwrwrQdrrVbo3jdzaWFcbhqdDg5d7yvdM2q3p1oTlrKC+q1E6MvxVUr3d9G9x3EJ1Y8L7r0Laswtjq+aCKrODRlhJOVBJ6pGscpNGHaeDWn2LJ6PciCycQXl7AUnk7mmzK1FclO8ewWqtWk0WhkE9UyqvewZrbR3HU7q4i9ENUneC9Amfc20BWVwIK5rP0IIABPxCN8K3s0zzp6fFLiw1Rel/oeYNRAM4OrysRGT0eT9xdxatfqrlQPXgYYepzaMZ9Ws+5uYVBlVjeHY1Bq+QJdK4SV4uG2Y2cyi+XWs+x0zUaznYAAKwAAAMa/5TOQzs1vypdjjMldfjC1+g1S8gqM0vKyLn43WrKVdEXWpliG4wutzOTnJuuZ4jS5lBTWsH9mZ4CslRcHrF/yM3v5s08mjlxTwmI4X5JaP0Em5pvjq9nsbiKtOjHlvh4m02tchNp1PD1J/olf2eRtj86NNLVyf9BPl4fDQjW+a+kf7ZqJZ3WeHUKtDlzzs33JcJcqdObuoq6fYh1ZU6UayppRk7ZamsJxqwbj1i19g31pfwyX4fD6j8spXOR4dPhkludipuSmPg0n3zBeZoiXR+wPzJgTLylYbF+jRnF2YX6XecOxTZmvVoy6MI26G1F5tC8Wi9N8M0wk7mjpBJAcgAABE1xU5R3TPInsEeSqR4ako7No1ENyp8eBhVWsG0+zETt+HRVXC1KMtG/5RxZxcJOEtU7FHX8Lq5SpPujrnmMJV5VeMul7P3PT9TNVAEtWIIFKytLiXU6NOXHBS3E6yvC+xfCTvFw2LG73icAANOYAAApUzhLscWXU7ks4tehxJasldPjR/gxS0Yub0epG8vDC1M8R+X7mi1RSv+UyVyx9jnFKlONaPBL29C5rSjxNr0ZI9N1rslTocLTnJz4fLfoK+IO+IUdopL6HUlCUVeSsczHr8eM/3RX2yNRyzk10aqwvhZU1qkrexhhaU6cvn/V0HUm9DCFSnKolB3tbQm27jNksLLhkntI9BLOJ56EWpVFtJnepTU6Sks7otZxvidYMhu8b7Ex2COcbexFX9dzF5MvF/Km+hWeTuDH8XT0ZV5SBPLsTPcIiOWRo8jHRmuYJ6ei7xTJMqLvC2xqHKzVAAARK1PL4yPDiai9f5PTnnvElbFS9UmaiGvCpfmR7Mx8TpcFZVFpNfdB4XNRrSTdrxHse6NTDtcceKOazH2PPHqMLU51CM+trPujy51/C6vmovuhR2AADKoaurCtGXLrJPsNClZcM+JdQ3h+OqBSnLjgpboubcwAAAPQ4k8pM7ZxamU37krp8bNdDei8/Yx/02o+YjeXhhaoirnTl2J2JnnBr0JXKeuUbU606eSMQI9Vm15y45OW4riabrUrLzRzX9o3zehpGm3rkN6Zyk1qko14vDSd/na4bdbsjD4apShxzVmx9Tw+HmpStf6spicW6kXGMeFX1lr9Pc1j24ZZd7jj16UnVqO9ouTd2dXAu9BR2yOXxNylGd202vWw9gOKLcZZZL7FrON7dDRgspP6hLUOqfsZd6I9Ykz0K3tNepZ6A+1Y7E6x7FI5MutWgVV6JmkfKZ9Gi8X9wzW9GVpW3GhCL4ZJj4ZznaAJWZAYScPxZWrRlvE7ZyfFl8tOXdGojmYeg8RU5cXbK+Z0o+FfvqfRCnhztil6pnoxaPK4mi8PWlT1S0foThavKrxn0vZ9mdTxWlxQjXXTJnCKPYgLYSrzcPGXW1n7DJhUGNeN4X2NiGrpp9QsuqjCTvBw2HDl4eXLrWfXI6hqLnOwAAVgHHrL8R+52Dk4hfivuSt4elzal50Y9DSn5l3I63w30LPNNFHoXJXBy4wlLRGkaX7jqWVrdBBqzaM59On9LVJqUYPlJX9TnQqVK8b1JNK7Xy5LL1OlKKlFxejMoUKNNWhFGcbJ6zSfCmuGjC6le77rdjsMI60eKtLJ6xj9NS7NqEtY+5vHPdZscytCGHcsrJP3Zjhqz58eNWvp/hv4rfijk7Wefqcul+bBrpb2VzpUkeikQ/L2JeaKrNWMPR+InlnszRmcs4+xaLbimC/VZ6SL6ST9ik9blm7xvsFqXkyE2vZ2CV8mR17oM1oPQfFBMR1Vxmg8mtgmXc23uSQAcgc/xRXw0ZbSOgJ+IR4sHP0s/uWDjYB2xUPf+D0x5bBZYqn3PUCitSmqtOVJ/qR5KUXGTi9VkewPP+J0eXX5i0mr+5YjXwurnKi+uaO0eVw1Xk14z6J59j1RKqABgyBKsuGpxLrmdWEuOCluhCvG8b7G2DneDhsWN3vHZsAA05oOZifzTpHOxa+e/YN4elOjNIP5l3/wy3LxeZl1pzoyy0KlloiVwarQVrRtLi3GloZ1Y8UO2ZcpuEJkAQcGgysZcE0yWUYgK2Kw1Wm4u7v0scxKK+WnGWf8A9qdN0VVinF8DWWQlWoWm7NyXc9ErFP0m5Uotq2XX0JiY4W6p8D6M262M13xv+UrqiIaNbMla90VXntugomEXeNiZaFIZNkX6WWcOwO2TRMdWvci142Ki60tsa0XadtzGLv7oE7NPYJO5Y6QEXurkhxBjiVxYaov+rNiJK8JRfVMQeVwztiKf/pHqzx8ZOElKOqdzsUfFFpXj7r/DViOwJ4+lzcM2tYZo3p1qVZXpyTMq0cRVThTahF6vVskV5c7tDxGiqcY1bppWbJh4XSXnk39hqGDw0NIJ98y2xFqeIo1sqcrsZ4XZFElHyq3Ym7MqiSTTiK4aXBWSfXIaEavy1LruG8O9x2QMo1oOKbeqA1tnjVxLFQbXGug6QVJdOHlcldOx2HCEtUmYSwlJ6XRNOnNlcutCzoTXld+5XhlFWkSua6krZhxIpZvQtwSfQbow5WeocuIxy5E8p9WZ4my/BHYjhS0Q3yl1YcqGxeJtz5i803ojs8uGyJtsOJtxqMKkW7xaTRd63OlON4tHLk82hY6/H+L6NFZZSTW5W+VyZu+e5G40ej9DLSRomn7oyetyk800eUk/Yn0KS8vYsmgl/ULK/oy71KLz23RPS4J6fou9NehoK4eWsRojllNUCeKp4iuuXTahDq+rG3JLV2KOrTXUu2dOXHwpfrn9EMR8Nw0dby7sYeIgtLso8S+iG2phWlPDUKTvCCT3NxB16j62KOpN6thqfHXRbS1ZR1aa6nOuFyNfzOvEQWl2UeIfRCyjN6I0VCrLoU44xLrzfoZOV3djMcHN+Z2N44Sms5ZjRzxnhD5vUDs2QF0z/T/iwABpyQBIAQBIAQBIAQAABAEgBAWJACrRzq2DqNt0ms+jOmBFl040cFiYp3aeyRT4fFJNSh2szuANLyrhxp14+aDyMpvhdm7HobGUqNKTvKKb9UTSzOxw1UU43RaNRWS2yO06NKSs4qxn8Jh+kErjS83KclxJml1Zjc8DCStBuJi8FVivlnxb3GjmxTtmizqSerZXkYi9lTf1RpHCYiXmtH7k06c8WVyLjscD+6X0No4SjHpfuNJ/SOZfYsoVJaI7CpU46JF7JaF4s35a5Kw1WXobRwb/AFM6AF0zc6UjhKa1NlRprRGoDTNtVUYrRFiQKiAJACAJAD//2Q=="/>

        <script type="text/javascript">
             {{if .SelectedGameID}}
             window.selectedGameID = "{{.SelectedGameID}}";
             {{end}}
             window.autogeneratedGameID = "{{.AutogeneratedGameID}}";
        </script>
    </head>
    <body>
		<div id="app">
		</div>
    </body>
</html>
`

type templateParameters struct {
	SelectedGameID      string
	AutogeneratedGameID string
}

func (s *Server) handleIndex(rw http.ResponseWriter, req *http.Request) {
	dir, id := filepath.Split(req.URL.Path)
	if dir != "" && dir != "/" {
		http.NotFound(rw, req)
		return
	}

	autogeneratedID := s.getAutogeneratedID()

	err := s.tpl.Execute(rw, templateParameters{
		SelectedGameID:      id,
		AutogeneratedGameID: autogeneratedID,
	})
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) getAutogeneratedID() string {
	const attemptsPerWordCount = 5

	s.mu.Lock()
	defer s.mu.Unlock()

	var words []string
	autogeneratedID := ""
	for i := 0; ; i++ {
		wordCount := 2 + i/attemptsPerWordCount

		words = words[:0]
		for j := 0; j < wordCount; j++ {
			w := s.gameIDWords[rand.Intn(len(s.gameIDWords))]
			words = append(words, w)
		}

		autogeneratedID = strings.Join(words, "-")
		if _, ok := s.games[autogeneratedID]; !ok {
			break
		}
	}
	return autogeneratedID
}
