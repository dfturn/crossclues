# Build frontend.
FROM node:12-alpine as frontend
COPY . /app
WORKDIR /app/frontend
RUN npm install -g parcel-bundler \
    && npm install
RUN sh build.sh

# Build backend.
FROM golang:1.14-alpine as backend
WORKDIR /app
COPY . .
RUN apk add gcc musl-dev
RUN go build ./cmd/crossclues/main.go

# Copy build artifacts from previous build stages (to remove files not necessary for
# deployment).
FROM alpine:3.11
WORKDIR /app
COPY --from=backend /app/main .
COPY --from=frontend /app/frontend/dist ./frontend/dist
COPY assets assets
EXPOSE 8080/tcp
CMD /app/main
