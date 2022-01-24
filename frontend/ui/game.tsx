import * as React from 'react';
import axios from 'axios';
import { Settings, SettingsButton, SettingsPanel } from '~/ui/settings';
import Timer from '~/ui/timer';
import { ClientJS } from 'clientjs';

export class Game extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      game: null,
      mounted: true,
      settings: Settings.load(),
      mode: 'game',
      playerID: new ClientJS().getFingerprint().toString(),
      gameOver: false,
      timerExpired: false,
    };
  }

  public extraClasses() {
    var classes = '';
    if (this.state.settings.colorBlind) {
      classes += ' color-blind';
    }
    if (this.state.settings.darkMode) {
      classes += ' dark-mode';
    }
    if (this.state.settings.fullscreen) {
      classes += ' full-screen';
    }
    return classes;
  }

  public handleKeyDown(e) {
    if (e.keyCode == 27) {
      this.setState({ mode: 'game' });
    }
  }

  public componentDidMount(prevProps, prevState) {
    window.addEventListener('keydown', this.handleKeyDown.bind(this));
    this.setDarkMode(prevProps, prevState);
    this.refresh();
  }

  public componentWillUnmount() {
    window.removeEventListener('keydown', this.handleKeyDown.bind(this));
    this.setState({ mounted: false });
  }

  public componentDidUpdate(prevProps, prevState) {
    this.setDarkMode(prevProps, prevState);
  }

  private setDarkMode(prevProps, prevState) {
    if (!prevState?.settings.darkMode && this.state.settings.darkMode) {
      document.body.classList.add('dark-mode');
    }
    if (prevState?.settings.darkMode && !this.state.settings.darkMode) {
      document.body.classList.remove('dark-mode');
    }
  }

  private refreshEog() {
    if (
      !this.state.gameOver &&
      this.state.game != null &&
      this.state.game.won
    ) {
      this.state.gameOver = true;
      this.endGame();
    }
  }

  /* Gets info about current score so screen readers can describe how many words
   * remain for each team. */
  private getScoreAriaLabel() {
    return this.remaining().toString() + ' words remaining.';
  }

  // Determines value of aria-disabled attribute to tell screen readers if word can be clicked.
  private cellDisabled(idx) {
    if (this.state.game.revealed[idx]) {
      return true;
    } else if (this.state.game.won) {
      return true;
    }
    return false;
  }

  // Gets info about word to assist screen readers with describing cell.
  private getCellAriaLabel(idx) {
    let ariaLabel = this.state.game.revealed[idx]
      ? 'revealed word'
      : 'hidden word';
    ariaLabel += '.';
    return ariaLabel;
  }

  public refresh() {
    if (!this.state.mounted) {
      return;
    }

    this.refreshEog();

    let state_id = '';
    if (this.state.game && this.state.game.state_id) {
      state_id = this.state.game.state_id;
    }

    axios
      .post('/game-state', {
        game_id: this.props.gameID,
        state_id: state_id,
        player_id: this.state.playerID,
      })
      .then(({ data }) => {
        this.setState((oldState) => {
          const stateToUpdate = { game: data };
          return stateToUpdate;
        });
      })
      .finally(() => {
        setTimeout(() => {
          this.refresh();
        }, 2000);
      });
  }

  public guess(e, idx) {
    e.preventDefault();
    if (this.state.game.revealed[idx]) {
      return; // ignore if already revealed
    }
    if (this.state.game.won) {
      return; // ignore if game is over
    }
    if (this.state.timerExpired) {
      return; // ignore if timer expired
    }

    axios
      .post('/guess', {
        game_id: this.state.game.id,
        index: idx,
        player_id: this.state.playerID,
      })
      .then(({ data }) => {
        this.setState({ game: data });
      })
      .finally(() => {
        setTimeout(() => {
          this.refresh();
        }, 2000);
      });
  }

  public discard(e, idx) {
    if (!e) var e = window.event;
    e.cancelBubble = true;
    if (e.stopPropagation) e.stopPropagation();

    e.preventDefault();
    if (this.state.game.won) {
      return; // ignore if game is over
    }
    if (this.state.timerExpired) {
      return; // ignore if timer expired
    }

    axios
      .post('/discard', {
        game_id: this.state.game.id,
        index: idx,
        player_id: this.state.playerID,
      })
      .then(({ data }) => {
        this.setState({ game: data });
      })
      .finally(() => {
        setTimeout(() => {
          this.refresh();
        }, 2000);
      });
  }

  public remaining() {
    var count = 0;
    for (var i = 0; i < this.state.game.revealed.length; i++) {
      if (this.state.game.revealed[i]) {
        count++;
      }
    }
    return count;
  }

  public getRowName(row) {
    const rows = ['A', 'B', 'C', 'D', 'E'];
    return rows[row];
  }

  public getColName(col) {
    const cols = ['1', '2', '3', '4', '5'];
    return cols[col];
  }

  public getIndexName(idx) {
    const boardSize = this.state.game.board_size;

    const row = Math.floor(idx / boardSize);
    const col = idx % boardSize;
    return this.getRowName(row) + this.getColName(col);
  }

  public toggleSettingsView(e) {
    if (e != null) {
      e.preventDefault();
    }
    if (this.state.mode == 'settings') {
      this.setState({ mode: 'game' });
    } else {
      this.setState({ mode: 'settings' });
    }
  }

  public toggleSetting(e, setting) {
    if (e != null) {
      e.preventDefault();
    }
    const vals = { ...this.state.settings };
    vals[setting] = !vals[setting];
    this.setState({ settings: vals });
    Settings.save(vals);
  }

  public nextGame(e) {
    if (e != null) {
      e.preventDefault();
    }
    // Ask for confirmation when current game hasn't finished
    let allowNextGame =
      this.state.game.won ||
      this.state.timerExpired ||
      confirm('Do you really want to start a new game?');
    if (!allowNextGame) {
      return;
    }

    axios
      .post('/next-game', {
        game_id: this.state.game.id,
        player_id: this.state.playerID,
        word_set: this.state.game.word_set,
        create_new: true,
        timer_duration_ms: this.state.game.timer_duration_ms,
        enforce_timer: this.state.game.enforce_timer,
        hand_size: this.state.game.hand_size,
        board_size: this.state.game.board_size,
      })
      .then(({ data }) => {
        this.setState({ game: data, gameOver: false, timerExpired: false });
      })
      .finally(() => {
        setTimeout(() => {
          this.refresh();
        }, 2000);
      });
  }

  public endGame() {
    const messages = [
      "Oh dear, clearly you don't understand each other at all.",
      'You have a basic understanding of how the other players think!',
      'Wow! You have a strong connection!',
      'A perfect score! You must be telepathically linked!',
    ];

    let boardSizeIndex = this.state.game.board_size - 3;
    let score = parseInt(this.state.game.score);

    const thresholds = [
      [0, 4, 6, 8],
      [0, 8, 12, 15],
      [0, 12, 17, 23],
    ];

    let threshold = thresholds[boardSizeIndex];

    var messageIndex = threshold.length - 1;
    for (var i = 0; i < threshold.length; i++) {
      let thresholdVal = threshold[i];
      if (score < thresholdVal) {
        messageIndex = i - 1;
        break;
      }
    }

    let message = messages[messageIndex] + '\nPlay another one?';

    let startNewGame = confirm(message);
    if (startNewGame) {
      this.nextGame(null);
    }
  }

  render() {
    if (!this.state.game) {
      return <p className="loading">Loading&hellip;</p>;
    }
    if (this.state.mode == 'settings') {
      return (
        <SettingsPanel
          toggleView={(e) => this.toggleSettingsView(e)}
          toggle={(e, setting) => this.toggleSetting(e, setting)}
          values={this.state.settings}
        />
      );
    }

    let statusClass;
    if (this.state.game.won) {
      statusClass = 'win';
    } else {
      statusClass = 'turn';
    }

    let numberWords = this.state.game.words.filter((w, idx) => {
      return idx < this.state.game.words.length / 2;
    });

    let letterWords = this.state.game.words.filter((w, idx) => {
      return idx >= this.state.game.words.length / 2;
    });

    let shareLink = (
      <div id="share">
        Send this link to friends:&nbsp;
        <a className="url" href={window.location.href}>
          {window.location.href}
        </a>
      </div>
    );

    const timer = !!this.state.game.timer_duration_ms && (
      <div id="timer">
        <Timer
          roundStartedAt={this.state.game.created_at}
          timerDurationMs={this.state.game.timer_duration_ms}
          handleExpiration={() => {
            if (this.state.game.enforce_timer && !this.state.timerExpired) {
              this.state.timerExpired = true;
              this.endGame();
            }
          }}
          freezeTimer={!!this.state.game.won}
        />
      </div>
    );

    const cardsInHand = Object.keys(this.state.game.player_cards);

    const cellPercentSize = 85 / (this.state.game.board_size + 1);
    const cellPercentString = cellPercentSize + '%';

    var cellStyle = {
      '--width': cellPercentString,
      '--height': cellPercentString,
    } as React.CSSProperties;

    const gridWidth = this.state.game.board_size + 1;
    const gridSpaces = gridWidth * gridWidth;

    return (
      <div id="game-view" className={'player' + this.extraClasses()}>
        <div id="infoContent">
          {shareLink}
          {timer}

          <div id="remaining" role="img">
            <span className={'remaining'}>
              {'Score: ' +
                this.remaining() +
                ' / ' +
                this.state.game.revealed.length}
            </span>
          </div>

          <button onClick={(e) => this.nextGame(e)} id="next-game-btn">
            Next game
          </button>

          <SettingsButton
            onClick={(e) => {
              this.toggleSettingsView(e);
            }}
          />
        </div>

        <div className={'board ' + statusClass} style={cellStyle}>
          {[...Array(gridSpaces)].map((x, i) => {
            const row = Math.floor(i / gridWidth);
            const col = i % gridWidth;

            if (i == 0) {
              return <div className={'header'} key={i}></div>;
            } else if (row == 0) {
              return (
                <div className={'header'} key={i}>
                  <span className="word">{numberWords[col - 1]}</span>
                  <span className="letter">{this.getColName(col - 1)}</span>
                </div>
              );
            } else if (col == 0) {
              return (
                <div className={'header'} key={i}>
                  <span className="word">{letterWords[row - 1]}</span>
                  <span className="letter">{this.getRowName(row - 1)}</span>
                </div>
              );
            } else {
              const revealedIdx = (row - 1) * gridWidth + (col - row);

              return (
                <div
                  key={i}
                  idx={revealedIdx}
                  className={
                    'cell ' +
                    (this.state.game.revealed[revealedIdx]
                      ? 'revealed'
                      : 'hidden-word')
                  }
                  onClick={(e) => this.guess(e, revealedIdx)}
                >
                  <span className="word" role="button">
                    {this.state.game.revealed[revealedIdx]
                      ? this.getIndexName(revealedIdx)
                      : revealedIdx in this.state.game.discards
                      ? 'X'
                      : ''}
                  </span>
                </div>
              );
            }
          })}
        </div>

        <div className={'cardsInHand'}>
          {cardsInHand.map((w, idx) => (
            <div
              key={idx}
              className={'cell revealed'}
              onClick={(e) => this.guess(e, parseInt(w))}
            >
              <span className="letter" role="button">
                {this.getIndexName(parseInt(w))}
              </span>

              <button
                className="close"
                onClick={(e) => this.discard(e, parseInt(w))}
              ></button>
            </div>
          ))}
        </div>
      </div>
    );
  }
}
