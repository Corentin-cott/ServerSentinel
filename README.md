# Server Sentinel

## Table of Contents

- [Description](#Goals)
- [Functionalities](#Functionalities)
- [How to install](#How to install)
- [How to set console triggers](#How to set console triggers)
- [Contributions](#Contributions)

## Description

A CLI Tool and Deamon made for managing tmux server sessions and player server data. Storing player connections and game data, while also reading "servers terminals" for triggers.
For performance reasons, the Daemon doesn't read the terminal directly, but the logs the tmux session creates. The daemon then read the file each 1 second (this can be changed in the config).

## Functionalities

Main functionalities :
- Get and store server player data in database
- Read logs of [tmux](https://doc.ubuntu-fr.org/tmux) game server sessions to listen to server consoles<br>**->** Do stuff when certain things appear in server console *(Sent message with a bot, extract and store data, ect...)*
- A few CLI commands : use serveursentinel to know more about these commands
- 

## How to install

Section not filled yet !

## How to set console triggers

Section not filled yet !

## Contributions

[Corentin COTTEREAU (Azertor/Cocow)](https://github.com/Corentin-cott)
