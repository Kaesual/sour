<p align="center">
    <a href="https://sourga.me" target="_blank">
        <img src="gh-assets/header.png" alt="Sour Cover Image">
    </a>
</p>

## What is this?

This is a fork of the excellent [original sour repository](https://github.com/cfoust/sour) made by cfoust. It has been made with the goal to embed sauerbraten as an in-community game experience into the Common Ground platform. I want to enable a Quake-like experience, and add additional features like Tournaments between communities.



Sour is a <a target="_blank" href="http://sauerbraten.org/">Cube 2: Sauerbraten</a> server that serves a fully-featured web-version of Sauerbraten (with support for mobile devices) in addition to accepting connections from the traditional, desktop version of the game. Sour is the easiest way to play Sauerbraten with your friends.

<a target="_blank" href="https://sourga.me/">Give it a try.</a>

## Installation

You can download an archive containing the Sour server and all necessary assets from [the releases page](https://github.com/cfoust/sour/releases). For now, only Linux and macOS are supported.

You can also install Sour via `brew`:

```bash
# Install the latest version:
brew install cfoust/taps/sour

# Or a specific one:
brew install cfoust/taps/sour@0.2.2
```

In addition to all of the base game assets, these archives only contain three maps: `complex`, `dust2`, and `turbine`.

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

## Running Sour

To run Sour, extract a release archive anywhere you wish, navigate to that directory, and run `./sour`. If you installed Sour with `brew`, just run `sour` in any terminal session.

Running `sour` will start a Sour server accessible to web clients at `http://0.0.0.0:1337` and to desktop clients on port 28785. In other words, you should be able to connect to the Sour server in the Sauerbraten desktop client by running `/connect localhost`.

By serving on `0.0.0.0` by default, the Sour server will be available to other devices on the local network at IP of the device running the Sour server.

### Building the web game with Docker (Emscripten)

If you prefer building in a containerized environment, a Dockerfile and helper script are provided:

```bash
# Build the image and compile the game into game/dist/game
./scripts/build-game-docker

# Optionally control output directory
GAME_OUTPUT_DIR=client/dist/game ./scripts/build-game-docker
```

This uses an Ubuntu base with Emscripten 3.1.8 (same as CI), mounts your checkout at `/workspace`, and runs `game/build` inside the container. Artifacts will appear under `game/dist/game` by default.

### Running the server in Docker

After building the game and client (and optionally assets), you can run the integrated server with:

```bash
# Default: serves on 0.0.0.0:1337
./scripts/serve-docker

# With a config file
./scripts/serve-docker dev.yaml

# Override bind address/port
WEB_ADDR=127.0.0.1 WEB_PORT=1337 ./scripts/serve-docker
```

The script mounts your workspace and runs `go run ./cmd/sour serve` inside the container using your UID/GID so no files are owned by root.

## Configuration

Sour is highly configurable. When run without arguments, `sour` defaults to running `sour serve` with the [default Sour configuration](https://github.com/cfoust/sour/blob/main/pkg/config/default.yaml). You change Sour's configuration by providing the path to a configuration file to `sour serve`:

```bash
sour serve config.yaml
```

Sour can be configured using `.yaml` or `.json` files; the structure is the same in both cases.

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

- dockerized the build pipeline for the game client as well as assets
- uses one docker helper container that compiles everything and can also serve the game server
- fixed a bug that prevented keyboard events to work in iframes



## Contributing

This repository is maintained by the Common Ground Team (I'm one of the founders) as an in-community gaming experience. Common Ground itself is a progressive web app and supports embedding custom games and plugins into Communities. If you're interested in the project, join our [Common Ground community on app.cg](https://app.cg/c/commonground/).

Besides Sour / Sauerbraten, I also made a Luanti (think "open source minecraft") game plugin available on app.cg. Like Sour, it is also a web assembly game with an original c++ codebase. You can find my [minetest-wasm repository here](https://github.com/Kaesual/minetest-wasm). You can play both games right in your browser, in the [Video Games community on app.cg](https://app.cg/c/videogames/).

The original repository was made by cfoust. You can join the community on [Discord](https://discord.gg/WP3EbYym4M) to chat with them and see how you can help out! Check out the [cfoust sour issues tab](https://github.com/cfoust/sour/issues) to get an idea of what needs doing.

## Inspiration

Some years ago I came across [BananaBread](https://github.com/kripken/BananaBread), which was a basic tech demo that used [Emscripten](https://emscripten.org/) to compile Sauerbraten for the web. The project was limited in scope and done at a time when bandwidth was a lot more precious.

## License

Each project that was forked into this repository has its own original license intact, though the glue code and subsequent modifications I have made are licensed according to the MIT license specified in `LICENSE`.
