# Overview
GoChat is a chat backend which exposes a RESTful API for state management and WebSocket connections for real-time messaging for both direct messages and group chats. Key features include creating chat rooms, getting message history, creating, updating and deleting messages, adding and removing users from chat rooms, and user sign-up and login with JWT-based authentication.
# How to run
In the root directory, run:  
`docker compose build`  
`docker compose up -d`  
`docker compose --profile migration run --rm migrations`

# HTTP request lifecycle
<img width="1422" height="716" alt="image" src="https://github.com/user-attachments/assets/f8058004-f17f-4859-a493-eef5e85f8764" />


# WebSocket connection lifecycle
<img width="1436" height="678" alt="image" src="https://github.com/user-attachments/assets/bd02ea61-e3f5-4913-a6cc-49ab55a00d57" />
Note that this is for a 1 on 1 conversation between client 1 and client 2.
