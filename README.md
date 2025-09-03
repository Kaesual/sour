<p align="center">
    <img src="gh-assets/sour-cg.png" alt="Sour on app.cg Cover Image">
</p>

## What is Sour?

This is a fork of the excellent [original sour repository](https://github.com/cfoust/sour) made by cfoust. It has been made with the goal to embed sauerbraten as an in-community game experience into the Common Ground platform. I want to enable a Quake-like experience, and I'm planning to add additional features like Tournaments between communities - stay tuned! 🚀

Sour is a <a target="_blank" href="http://sauerbraten.org/">Cube 2: Sauerbraten</a> server that serves a fully-featured web-version of Sauerbraten (with support for mobile devices) in 
addition to accepting connections from the traditional, desktop version of the game. Sour is the easiest way to play Sauerbraten with your friends.

There's multiple deployments of this game available:

- The original version made and hosted by cfoust, available on [sourga.me](https://sourga.me)
- The in-community version on app.cg, in the [Common Games community](https://app.cg/c/commongames)
- A standalone version of the Common Ground version [here](https://embed.commonground.cg/sour/)

## The Common Games Collection

This repository is part of a broader effort to build a collection of Open Source games, which I call the [Common Games Collection](https://github.com/Kaesual/common-games-collection). Many such games exist, but are often cumbersome to build and thereby restricted to experts. I'm trying to build a unified collection, and make sure that all the games

- have a proper dockerized build pipeline (simple to run on any OS)
- can generate docker images ready for self hosting
- can easily be hosted on any path behind a reverse nginx proxy (for https support and structure)
- can be run in iframes, since this is required for my use case

My idea is that as a collective, we can build a collection of great games where the community can *focus on modifying the games*, and knowledge about the (sometimes delicate) process of *converting a game to web assembly* can be shared, too. This way, it becomes easier to add more games over time.

## Updates in this fork

- The build is now fully dockerized, completely in userspace
- The build generates a ready-to-host docker image with all assets. The image is also available [here](https://hub.docker.com/r/janhan/sour), but it currently lacks the config file (see `scripts/run-serve-image` for how to do it).
- Fixed an issue that prevented keyboard events from being picked up when running in iframes
- Made hosting possible under a relative path when combined with a reverse nginx proxy + rewrite rule
- And some more I probably forgot :D

The docker build is only tested on linux, with docker, but I tried to make it podman compatible.

## Local Development and Deployment

### Prerequisite: Git LFS

This repository stores large binary assets (textures, images, etc.) in Git LFS. After cloning, ensure LFS is installed and fetch objects, or some files will be tiny pointer stubs that fail at runtime.

Brief setup:

```bash
# Ubuntu/Debian
sudo apt install git-lfs

# macOS (Homebrew)
brew install git-lfs

# One‑time init, then pull LFS content
git lfs install
git lfs pull
```

### Building the web game with Docker (Emscripten)

A Dockerfile and helper script for building in userspace are provided. To avoid issues, make sure to check out the repository with the same user who will run the build. When running the build with docker, this user also needs to be in the docker group. For podman this is not necessary.

```bash
# Build everything and put it into a nice new docker image, ready to host
./scripts/build-all
```

This uses an Ubuntu base with Emscripten 3.1.8 (same as CI), mounts your checkout at `/workspace`, and runs the build scripts in the container. It creates a new docker image called `sour-serve:latest` by default. `build-all` is just a wrapper for the following build scripts:

```bash
scripts/build-docker-image # builds the builder docker image
scripts/build-assets
scripts/build-game
scripts/build-proxy
scripts/build-web
scripts/build-serve-image
```

All steps can be run independently, e.g. if you only updated the web interface, you can run `scripts/build-web && scripts/build-serve-image` to update the image. Some assets are downloaded and cached during the first asset build, so it takes longer the first time. After that, building assets runs quite fast.

### Running the server in Docker

After building, you can run the integrated server locally with:

```bash
# Default: serves on 0.0.0.0:1337
./scripts/run-serve-image

# Override bind address/port
WEB_ADDR=127.0.0.1 WEB_PORT=1337 ./scripts/run-serve-image
```

The script mounts your workspace and runs `go run ./cmd/sour serve` inside the container using your UID/GID so no files are owned by root. There's also the older `scripts/serve` that I used before the docker images, should still work if you don't want to re-build the container every time, but must be restarted after making updates.

## Configuration

Sour is highly configurable. When run without arguments, `sour` defaults to running `sour serve` with the [default Sour configuration](https://github.com/Kaesual/sour/blob/main/pkg/config/default.yaml). By default, the `run-serve-image` script mounts the `dev.auto.yaml` from this repository folder. You can make changes there or mount your own config file.

Sour can be configured using `.yaml` or `.json` files; the structure is the same in both cases.

Warning: The section below is from the original readme and hasn't been updated yet, it will probably need a docker command instead as the docker image is where the system requirements are installed.

To print the default configuration to standard output, run `sour config`:

```bash
sour config > config.yaml
```

Sour also supports merging configurations together.

```bash
sour serve config_a.yaml some_path/config_b.json config_c.yaml
```

These configurations are merged from left to right using [CUE](https://cuelang.org/docs/). In other words, configurations are evaluated in order from left to right. CUE merges JSON data by overwriting values (if they're scalar, such as strings, booleans, and numbers) or combining values (if they're arrays). In effect, this means that configurations can specify values for only a subset of properties without problems.

## Goals

- **Modernize Sauerbraten.** The gaming landscape has changed. Provide a modern multiplayer experience with matchmaking, private games, rankings, and seamless collaboration on maps. Make as much of this functionality available to the unmodified desktop game as possible.
- **Preserve the experience of playing the original game.** While it is possible that Sour may someday support arbitrary game modes, assets, clients, and server code, the vanilla game experience should still be available.
- **Be the best example of a cross-platform, open-source FPS.** Deployment of Sour on your own infrastructure with whatever configuration you like should be easy. Every aspect of Sour should be configurable.

Note: The goals above are originally from cfoust. My own goal to add would be embedding Sour into app.cg in a way that there's cross-community gaming activity.

## Architecture

Here is a high level description of the repository's contents:

- `pkg` and `cmd`: All Go code used in Sour and its services.
  - `cmd/sourdump`: A Go program that calculates the minimum list of files necessary for the game to load a given map.
  - `cmd/sour`: The Sour game server, which provides a number of services to web clients:
    - Gives clients both on the web and desktop client access to game servers managed by Sour.
- `game`: All of the Cube 2 code and Emscripten compilation scripts. Originally this was a fork of [BananaBread](https://github.com/kripken/BananaBread), kripken's original attempt at compiling Sauerbraten for the web. Since then I have upgraded the game to the newest mainline version several times and moved to WebGL2.
- `client`: A React web application that uses the compiled Sauerbraten game found in `game`, pulls assets, and proxies all server communication over a WebSocket.
- `assets`: Scripts for building web-compatible game assets. This is an extremely complicated topic and easily the most difficult aspect of shipping Sauerbraten to the web. Check out this [section's README](assets/README.md) for more information.

**Updates in this fork**

- `scripts`: dockerized the build pipeline for the game client as well as assets
- uses one docker helper container that compiles everything and can also serve the game server
- fixed a bug that prevented keyboard events to work in iframes
- provide several simple build scripts for different steps of the process

## Inspiration

Original text from cfoust

Some years ago I came across [BananaBread](https://github.com/kripken/BananaBread), which was a basic tech demo that used [Emscripten](https://emscripten.org/) to compile Sauerbraten for the web. The project was limited in scope and done at a time when bandwidth was a lot more precious.

## License

Original text from cfoust

Each project that was forked into this repository has its own original license intact, though the glue code and subsequent modifications I have made are licensed according to the MIT license specified in `LICENSE`.
