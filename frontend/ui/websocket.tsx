export class Websocket {
  constructor() {
    this.websocket = null;
  }

  public connect(gameID, playerID) {
    this.close();

    var getUrl = window.location;
    let wsProtocol = getUrl.protocol.endsWith('s:') ? 'wss://' : 'ws://';
    let wsPath =
      wsProtocol + getUrl.host + '/websocket/' + gameID + '/' + playerID;

    this.websocket = new WebSocket(wsPath);
    return this.websocket;
  }

  public close() {
    if (this.websocket) {
      this.websocket.close();
    }
    this.websocket = null;
  }
}

const websocket = new Websocket();
export default websocket;
