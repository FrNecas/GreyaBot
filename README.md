# GreyaBot

A discord bot created for managing a server used by the community around a Twitch streamer.

## Features

- A welcome customizable message sent to each new user.
- User verification requiring agreeing to the rules of a server.
- Language filter that can remove certain phrases. An example use-case of this is avoiding disclosure of personal information.
- Reacting to commands predefined in the configuration file.
- A twitch module used for connecting to the Twitch webhook API and receiving updates about the stream. The updates are sent as an embedded discord message.

## Building and running

The project can be built and run using the go tool ecosystem. Either `go build` or `go run main.go` can be used. The project has one external dependency - [discordgo](https://github.com/bwmarrin/discordgo). In order to run the bot, a file called `config.json` that contains a valid configuration must be present in the current working directory. 

An alternative, especially useful for deploying, is using a container. The application utilizes multi-stage build to reduce the final image size to a few MB. The image can either be built using the provided Containerfile or downloaded from [docker hub](https://hub.docker.com/r/fnecas/greyabot). An example how to run the container:

```
podman run -v ./config.json:/config.json docker.io/fnecas/greyabot
```

The twitch integration runs a web server, hence exposing a port is required. To make this and volume mapping more convenient, `docker-compose.yml` is provided; it is sufficient to run:

```
podman-compose up
```

## Configuring

The whole bot is configured through `config.json`. The file must be a valid JSON and be present in current working directory when running the bot. A minimal configuration requires a discord API token. `config.json.example` contains all possible fields and can be used as a guideline when creating a configuration. For more information and documentation about the config, refer to `config/config.go`.

