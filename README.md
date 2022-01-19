# crossclues

CrossClues implements a web app for generating and displaying boards for the <a href="https://boardgamegeek.com/boardgame/300753/cross-clues">Cross Clues</a> board game. Generated boards are shareable and will update as words are revealed.

A hosted version of the app is available at [www.crossclues-t2pg6luaoa-uw.a.run.app](https://crossclues-t2pg6luaoa-uw.a.run.app).

# Credit

This app is forked and heavily based on the [codenames app](https://github.com/jbowens/codenames) developed by jbowens.

## Building

The app requires a [Go](https://golang.org/) toolchain, node.js and [parcel](https://parceljs.org/) to build. Once you have those setup, build the application Go binary with:

```
go build cmd/crossclues/main.go && ./main
```

Then from the frontend directory, install the node modules:

```
npm install
```

and start the app (listens to changes)

```
npm start
```

or build the app

```
npm run build
```

### Docker

Alternatively, the repository includes a Dockerfile for building a docker image of this app.

```
docker build . -t crossclues:latest
```

The following command will launch the docker image:

```
docker run --name crossclues_server --rm -p 8080:8080 -d crossclues
```

The following command will kill the docker instance:

```
docker stop crossclues_server
```
