# Discussion
a) discuss, whether you are going to use server-side streaming, client-side streaming, or bidirectional streaming?
- We are going to use server-side streaming to enable the client to receive a continous stream of messages while they are connected, and this stream will consist of chats and join notifications. This way users can have full conversations on chitchat and also be notified when other users connect. 

b) describe your system architecture - do you have a server-client architecture, peer-to-peer, or something else?
- We have a server-client architecture where clients request to connect to the server, post chats, and disconnect. The server serves the clients their received messages.

c) describe what RPC methods are implemented, of what type, and what  messages types are used for communication
- We have implemented three message types User, Message, and Nothing. User contains the username. Message contains the username of the sender, the content of the chat, and the time when the sender sent the chat. Nothing is a placeholder for nothing – an empty object. We have implemented three rpc methods. The first one is PostMessage which is unary. The client sends a Message to the server and when the chat is posted they receive Nothing. The second one is Connect which is server-side streaming. The client sends their user information User and when they are connected they receive a stream of Message. The third rpc is Disconnect which is also unary. The client sends their user information User and receive Nothing when they are disconnected from the server.

d) describe how you have implemented the calculation of the timestamps
Every Message and User also carries a Lamport timestamp. The server and each client keep their own internal integer clock, starting at 0.
1. Before a client sends a request (Connect or PostMessage) it increments its internal clock and attaches the value.
2. When the server receives a request it updates its internal clock as: clock = max(clock, received) + 1. Both the server and the client have a tick() helper function for this. A client leaving closes its stream. The server detects that an increments its clock as a local event, since the client doesn't send a timestamp when cancelling the terminal session.
3. The server includes the updated value of its internal clock for each broadcast message, so all clients receive the same timestamp (logical time). When a client receives a broadcast it updates its own clock as: clock = max(clock, received) + 1.

e) provide a diagram, that traces a sequence of RPC calls together with the Lamport timestamps, that corresponds to a chosen sequence of interactions: Client X joins, Client X Publishes, ..., Client X leaves.

Client X             Server              Client Y
   |                    |                    |
   |---- Connect ------>|                    |
   |<--- join notice ---|---- join notice -->|
   |                    |                    | der mangler timestamps
   |--- PostMessage --->|                    | 
   |                    |---- message ------>|
   |                    |                    |
   |---- disconnect --->|                    |