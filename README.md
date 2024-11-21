# Kinozal-Bot

Kinozal-Bot is a Telegram bot that allows users to search for, download, and manage torrents from Kinozal.tv. The bot integrates with Telegram, Kinozal, and Transmission for convenient torrent downloading and management directly from the chat interface.

## Table of Contents
1. [Project Description](#project-description)
2. [Installation](#installation)
   - [Using Binary Executables](#using-binary-executables)
   - [Using Docker](#using-docker)
3. [Commands](#commands)
4. [Usage](#usage)
5. [License](#license)

## Project Description
Kinozal-Bot allows you to:
- Search for torrents on Kinozal.tv.
- Download torrents directly to specified directories.
- Manage users: add and remove users who are allowed to use the bot.
- Integrate with Transmission for torrent management.
- Works on Linux, Windows, and macOS.

## Installation

### Using Binary Executables

1. Download the compiled binary for your operating system:
   - [Kinozal-Bot for macOS](#)
   - [Kinozal-Bot for Windows](#)
   - [Kinozal-Bot for Linux](#)

2. Extract the file if it is in an archive.

3. Ensure that the `.env` file is located in the same directory. If it is missing, create a `.env` file with the contents as shown below.
 Example `.env` file:
    ```env
    TG_TOKEN=<PASSWORD>:<PASSWORD>
    BOT_ADMIN_ID=<1234567890>
    BOT_ALLOWED_USERS=<1234567890,1234567891>
    KZ_ADDR="kinozal.tv"
    KZ_USER=<username>
    KZ_PASS=<PASSWORD>
    TRANS_ADDR=<http://localhost:9091/transmission/rpc>
    TRANS_USER=user
    TRANS_PASS=<PASSWORD>
    FILMS_FOLDER=</path/to/films>
    SERIES_FOLDER=</path/to/series>
    AUDIOBOOKS_FOLDER=</path/to/audiobooks > 
    ```

4. Run the binary:
   - On macOS or Linux:
     ```bash
     ./kinozal-bot-darwin
     ```
   - On Windows:
     ```bash
     kinozal-bot.exe
     ```

### Using Docker

If you prefer to use Docker, follow these steps:

1. Ensure Docker is installed on your machine. If not, download and install it from the [official Docker website](https://www.docker.com/get-started).

2. Create a `docker-compose.yml` file (if it doesn’t exist) with the following content:

   ```yaml
   version: "3.8"

   services:
     bot:
       image: dzarlax/kinozalbot:latest
       container_name: kinozal_bot
       environment:
         TG_TOKEN: ${TG_TOKEN}  # Secret Telegram token
         BOT_ADMIN_ID: ${BOT_ADMIN_ID}  # Admin ID
         BOT_ALLOWED_USERS: ${BOT_ALLOWED_USERS}  # Allowed user IDs
         KZ_ADDR: "kinozal.tv"  # Kinozal address
         KZ_USER: ${KZ_USER}  # Kinozal username
         KZ_PASS: ${KZ_PASS}  # Kinozal password
         TRANS_ADDR: ${TRANS_ADDR}  # Transmission address
         TRANS_USER: ${TRANS_USER}  # Transmission username
         TRANS_PASS: ${TRANS_PASS}  # Transmission password
         FILMS_FOLDER: ${FILMS_FOLDER}   # Films folder
         SERIES_FOLDER: ${SERIES_FOLDER}   # Series folder
         AUDIOBOOKS_FOLDER: ${AUDIOBOOKS_FOLDER}    # Audiobooks folder
         
       restart: unless-stopped
    ```
   

   Replace the values in `${}` with your own credentials or create a `.env` file with these values near the `docker-compose.yml` file.

   Example `.env` file:
    ```env
    TG_TOKEN=<PASSWORD>:<PASSWORD>
    BOT_ADMIN_ID=1234567890
    BOT_ALLOWED_USERS=1234567890,1234567891
    KZ_ADDR="kinozal.tv"
    KZ_USER=username
    KZ_PASS=<PASSWORD>
    TRANS_ADDR=http://localhost:9091/transmission/rpc
    TRANS_USER=user
    TRANS_PASS=<PASSWORD>
    FILMS_FOLDER=/path/to/films
    SERIES_FOLDER=/path/to/series
    AUDIOBOOKS_FOLDER=/path/to/audiobooks  
    ```

3. Run `docker-compose up -d` to start the container.

4. To stop the container, run `docker-compose down`.

	5.	The bot will be running inside the container and accessible via Telegram.

## Commands

Once the bot is up and running, you can use the following commands in Telegram:
```
	•	/start: Start the bot and receive a welcome message.
	•	/find [query]: Search for torrents on Kinozal.tv by name.
	•	/adduser [user_id]: Add a user to the allowed list (admins only).
	•	/removeuser [user_id]: Remove a user from the allowed list (admins only).
	•	/listusers: Display all allowed users (admins only).
	•	/help: Get a list of available commands.
   ```

## Usage

### Searching and Downloading Torrents

	1.	Use the /find [query] command to search for torrents on Kinozal.tv.
	2.	The bot will display a list of results with download buttons.
	3.	Select a torrent, and the bot will prompt you to choose a folder for downloading (e.g., Films, Series, Audiobooks).

### User Management

	1.	Administrators can manage bot access with the commands /adduser, /removeuser, and /listusers.
	2.	Only users with admin privileges can use these commands.

## License

This project is licensed under the MIT License.