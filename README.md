# Overview
GoChat is a chat backend which exposes a RESTful API for state management and WebSocket connections for real-time messaging for both direct messages and group chats.  
Key features include creating chat rooms, getting message history, creating, updating and deleting messages, adding and removing users from chat rooms, and user sign-up and login with JWT-based authentication.

# Tech Stack
- Go
- PostgreSQL
- Redis
- Docker
- pgx
- golang-migrate
- Chi
- WebSockets
- bcrypt
- JWT

# Architecture
The application has two main communication mechanisms between the client and the server.  
- REST, which deals with requests to a change in the application state. This includes creating users, logging in, creating rooms, adding members, retrieving message history, etc.  
- WebSockets, which provide the persistent connection through which the server and client can exchange events without repeatedly sending or listening for HTTP requests (polling). This includes sending/editing/deleting messages, and receiving these new changes.  

The WebSocket connection begins as an ordinary HTTP request before it is upgraded. Authentication will also occur before this step and the user ID of the request is stored in the context to create the Client object.  

The client is a server-side representation of an active WebSocket connection, and it contains the network connection itself, as well as the Hub (for broadcasting messages and handling membership), the room ID, the user ID, and an outgoing message channel to send the message to the client's `writePump`.  

The client has a `send` channel so only one goroutine (the `writePump`) is writing to a particular WebSocket connection. The Hub will send a message to this channel, and so the message flow for the `writePump` is from server to client.  

The `readPump` is responsible for reading WebSocket messages and then sending it to the hub's `broadcast` channel where it is broadcast to every connected user. It will also persist the messages in PostgreSQL so message history can be obtained. The message flow for the `readPump` is from client to server.  

The `Hub` manages the collection of currently connected WebSocket clients. This includes a map, where the key is the room ID and the value is a set of clients (members) which are currently connected to that room. It also has a global set of clients which are connected to any room, to easily check if a client is connected when trying to deregister them.  
The Hub's map is shared state, so concurrent access must be controlled. This is why, instead of using mutexes, I made sure only one goroutine owns the map, and other goroutines can send messages via channels to the hub goroutine to register, deregister, or broadcast a message. This follows *“Don't communicate by sharing memory; share memory by communicating"*.

**PostgreSQL represents the persistent state whereas the WebSocket Hub represents only the connections that are currently live.**

## HTTP request lifecycle
<img width="1422" height="716" alt="image" src="https://github.com/user-attachments/assets/f8058004-f17f-4859-a493-eef5e85f8764" />


## WebSocket connection lifecycle
<img width="1436" height="678" alt="image" src="https://github.com/user-attachments/assets/bd02ea61-e3f5-4913-a6cc-49ab55a00d57" />
Note that this is for a 1 on 1 conversation between client 1 and client 2.

# How to run
In the root directory, run:  
`docker compose build`  
`docker compose up -d`  
`docker compose --profile migration run --rm migrations`
