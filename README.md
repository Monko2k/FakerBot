
# FakerBot

Bootleg self-hostable Mikuia clone/np bot

I made this for my own stream, but you are free to use it if you want to




## Features

- Twitch/Bancho beatmap linker w/ mod support
- !np command
- !skin command



## Setup

Fill out `config.json` with the following values:

`TwitchUser`: your Twitch channel name

`TwitchPass`: your Twitch IRC authentication token (format is `oauth:<token>` where `<token>` is the token)

(Follow [this guide](https://dev.twitch.tv/docs/irc/guide) to get both of these)

`BanchoUser`: your osu! username

`BanchoPass`: your Bancho IRC authentication token 

(Follow [this guide](https://osu.ppy.sh/wiki/en/Internet_Relay_Chat) to get both of these) 

`GosuPort`: change this if needed (i.e. you changed the port the Gosumemory websocket runs on)




(Note: config file must be in same directory as the main executable)
    
## Acknowledgements

 - [flesnuk for his oppai port](https://github.com/flesnuk/oppai5)
 - [Francesco149 for the thing that flesnuk ported](https://github.com/Francesco149/oppai-ng)


Send feature requests/bug reports to Monko2k#3672 on discord

