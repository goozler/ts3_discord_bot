# TeamSpeak3 Discord Bot
[![GitHub release](https://img.shields.io/github/release/goozler/ts3_discord_bot.svg?style=flat-square)](https://github.com/goozler/ts3_discord_bot/releases) [![Docker Build Status](https://img.shields.io/docker/build/goozler/ts3_discord_bot.svg?style=flat-square)](https://hub.docker.com/r/goozler/ts3_discord_bot/) [![Travis branch](https://img.shields.io/travis/goozler/ts3_discord_bot/master.svg?style=flat-square)](https://travis-ci.org/goozler/ts3_discord_bot)

![](https://github.com/goozler/ts3_discord_bot/blob/master/screenshots/discord.jpg?raw=1)

Config Variables
------
You need to provide these are variables to configure the script
- ```TS3_DISCORD_BOT_HOST``` - hostname of the TS3 server
- ```TS3_DISCORD_BOT_PORT``` - port of the TS3 server for ServerQueries (10011)
- ```TS3_DISCORD_BOT_LOGIN``` - login for ServerQueries (see below how to get it)
- ```TS3_DISCORD_BOT_PASSWORD``` - password for ServerQueries (see below how to get it)
- ```TS3_DISCORD_BOT_WEBHOOK_URL``` - URL for a Discord webhook with `https://` prefix
- ```TS3_DISCORD_BOT_TIMEZONE``` - timezone for the imestamp (Europe/Samara)

ServerQuery credentials
------
You can retrieve them via the TS3 Client and a user with admin privileges

![](https://github.com/goozler/ts3_discord_bot/blob/master/screenshots/teamspeak_settings.jpg?raw=1)

Wait when the TS3 server is ready
------
Sometimes you need to wait until your server begins to receive requests. For this you can use `wait-for` script. An example of usage is in the `docker-compose.yml` below.

An example of Docker Compose config with a TS3 server
------
```yaml
version: '3'
services:
  ts3_server:
    image: <teamspeak3_image>
    ports:
      - 9987:9987/udp
      - 30033:30033

  discord_bot:
    image: goozler/ts3_discord_bot
    depends_on:
      - ts3_server
    command: ["./wait-for", "--timeout=5", "ts3_server:10011", "--", "./ts3_discord_bot"]
    environment:
      - TS3_DISCORD_BOT_WEBHOOK_URL=https://discordapp.com/api/webhooks/<webhook_id>
      - TS3_DISCORD_BOT_TIMEZONE=Europe/Samara
      - TS3_DISCORD_BOT_LOGIN=discord_bot
      - TS3_DISCORD_BOT_PASSWORD=34mb0hUB
      - TS3_DISCORD_BOT_HOST=ts3_server
      - TS3_DISCORD_BOT_PORT=10011
```
