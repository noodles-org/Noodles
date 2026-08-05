# Soba

Soba is a custom Discord bot that allows users to fetch the game server's external IP address via the `/get-server-ip` slash command.

## Deployment

- **Image:** `mephalrith/soba-discord-bot:v0.1`
- **Replicas:** 1
- **Namespace:** `soba`

Environment variables:

| Variable   | Source          | Description                              |
|------------|-----------------|------------------------------------------|
| `TOKEN`    | `soba-secret`   | Discord bot token                        |
| `GUILD_ID` | `soba-secret`   | Discord server (guild) ID                |
| `GAMES`    | Inline          | Game-to-port mapping for IP lookups      |

The `GAMES` variable is configured as `Enshrouded=15637, Satisfactory=7777, Valheim=2456`.

## Updating

The Soba bot image is updated via:
```
make update-soba-version
```
This rebuilds the Docker image, pushes it, and restarts the deployment.

## Source

The bot source code lives in `gameserver_deployment/infra/k8s/soba/soba-bot/` and is written in Go.
