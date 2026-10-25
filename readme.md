# Mandatory Exersice 3
Mikkel Hedeggard <mivh@itu.dk>
Hanibal Munk  <hmuk@itu.dk>
Abdhalla Jawher Alhamedi <aalh@itu.dk>

Github: https://github.com/Honningbolden/mandatory-exercise-3

All commands are run from the project root.

## How to run

1. Start the server:
   go run ./server

2. In a new terminal, start a client with a username:
   go run ./client <username>
   Repeat this in separate terminals to add more clients. Each username must be unique.

3. Type a message in a client terminal and press Enter to send it.
   Messages can be at most 128 characters.

4. To leave the chat, press Ctrl+C in a client terminal.
   The other clients are notified that the user left.

5. To stop the server, press Ctrl+C in the server terminal.
