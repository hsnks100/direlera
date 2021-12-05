# direlera

BUILD  
```
go build
./direlera
```

# todo
game list update... & user info update.

기존 유저 목록들 정보 이상하게 나옴. 왜 그러지?

방닫을 때 방 정리하기.


# Special Thank you

서영철(osio)


# wireshark

filter: ip.addr == 220.85.140.51
# packet structure

```
N
   seq1[2] length[2] msg_type [DATA]
   seq0[2] length[2] msg_type [DATA] ...

   length 는 msg_type 을 포함한 길이.
```
# packet spec

```
' 0x01 = User Quit
' Client Request:
' NB : Empty String [00]
' 2B : 0xFF
' NB : Message
'
' Server Notification:
' NB : Username
' 2B : UserID
' NB : Message
'
' 0x02 = User joined
' Server Notification:
' NB : Username
' 2B : UserID
' 4B : Ping
' 1B : Connection Type (6=Bad, 5=Low, 4=Average, 3=Good, 2=Excellent, &
1=LAN)
'
' 0x03 = User Login Information
' Client Notification
' NB : Username
' NB : Emulator Name
' 1B : Connection Type (6=Bad, 5=Low, 4=Average, 3=Good, 2=Excellent, &
1=LAN)
'
' 0x04 = Server Status
' Server Notification:
' NB : Empty String [00]
' 4B : Number of Users in Server (not including you)
' 4B : Number of Games in Server
' NB : List of Users
' NB : Username
' 4B : Ping
' 1B : Connection Type (6=Bad, 5=Low, 4=Average, 3=Good,
2=Excellent, & 1=LAN)
' 2B : UserID
' 1B : Player Status (0=Playing, 1=Idle)
' NB : List of Games
' NB : Game Name
' 4B : GameID
' NB : Emulator Name
' NB : Owner of Room
' NB : Number of Players/Maximum Players [MvC: 2/2, MvC4P: 4/2]
' 1B : Game Status (0=Waiting, 1=Playing, 2=Netsync)
'
' 0x05 = Server to Client ACK
' Server Notification:
' NB : Empty String [00]
' 4B : 00
' 4B : 01
' 4B : 02
' 4B : 03
'
' 0x06 = Client to Server ACK
' Client Notification:
' NB : Empty String [00]
' 4B : 00
' 4B : 01
' 4B : 02
' 4B : 03
'
' 0x07 = Global Chat
' Client Request:
' NB : Empty String [00]
' NB : Message
'
' Server Notification:
' NB : Username
' NB : Message
'
' 0x08 = Game Chat
' Client Request:
' NB : Empty String [00]
' NB : Message
'
' Server Notification:
' NB : Username
' NB : Message
'
' 0x09 = Client Keep Alive
' Client Request:
' NB : Empty String [00]
'
' 0x0A = Create Game
' Client Request:
' NB : Empty String [00]
' NB : Game Name
' NB : Empty String [00]
' 4B : 0xFF
'
' Server Notification:
' NB : Username
' NB : Game Name
' NB : Emulator Name
' 4B : GameID
'
' 0x0B = Quit Game
' Client Request:
' NB : Empty String [00]
' 2B : 0xFF
'
' Server Notification:
' NB : Username
' 2B : UserID
'
' 0x0C = Join Game
' Client Request:
' NB : Empty String [00]
' 4B : GameID
' NB : Empty String [00]
' 4B : 0x00
' 2B : 0xFF
' 1B : Connection Type (6=Bad, 5=Low, 4=Average, 3=Good, 2=Excellent, &
1=LAN)
'
' Server Notification:
' NB : Empty String [00]
' 4B : Pointer to Game on Server Side [client has no use for this...]
' NB : Username
' 4B : Ping
' 2B : UserID
' 1B : Connection Type (6=Bad, 5=Low, 4=Average, 3=Good, 2=Excellent, &
1=LAN)
'
' 0x0D = Player Information
' Server Notification:
' NB : Empty String [00]
' 4B : Number of Users in Room [not including you]
' NB : Username
' 4B : Ping
' 2B : UserID
' 1B : Connection Type (6=Bad, 5=Low, 4=Average, 3=Good, 2=Excellent, &
1=LAN)
'
' 0x0E = Update Game Status
' Server Notification:
' NB : Empty String [00]
' 4B : GameID
' 1B : Game Status (0=Waiting, 1=Playing, 2=Netsync)
' 1B : Number of Players in Room
' 1B : Maximum Players
'
' 0x0F = Kick User from Game
' Client Request:
' NB : Empty String [00]
' 2B : UserID
'
' 0x10 = Close game
' Server Notification:
' NB : Empty String [00]
' 4B : GameID
'
' 0x11 = Start Game
' Client Request:
' NB : Empty String [00]
' 2B : 0xFF
' 1B : 0xFF
' 1B : 0xFF
'
' Server Notification:
' NB : Empty String [00]
' 2B : Frame Delay (eg. (connectionType * (frameDelay + 1) <-Block on
that frame
' 1B : Your Player Number (eg. if you're player 1 or 2...)
' 1B : Total Players
'
' 0x12 = Game Data
' Client Request:
' NB : Empty String [00]
' 2B : Length of Game Data
' NB : Game Data
' *eg). MAME32K 0.64 = 2 Bytes/Input _
' Connection Type = (3=Good), so...3 * 2 = 6 Bytes for 1 Player's
Input)*
'
' Server Notification:
' NB : Empty String [00]
' 2B : Length of Game Data
' NB : Game Data
' *Using same example from above...If both players are on 3=Good
Connection Type _
' and there are 2 Players, then the Total size of the incoming data
should be: _
' 3 * 2 = 6 Bytes...6 Bytes * 2 Players = 12 Bytes*
'
' 0x13 = Game Cache
' Client Request:
' NB : Empty String [00]
' 1B : Cache Position
' *256 Slots [0 to 255]. Oldest to Newest. When cache is full add
new _
' entry at position 255 and shift all the old entries down knocking
off the oldest. _
' Search cache for matching data before you send. If found, send
that cache position, _
' otherwise issue a Game Data Send [0x12]. When server sends a game
Data Notify to you, _
' search for matching cache data, if not found add it to a new
position*
'
' Server Notification:
' NB : Empty String [00]
' 1B : Cache Position
' *Uses same cache procedure as above.*
'
' 0x14 = Drop Game
' Client Request:
' NB : Empty String [00]
' 1B : 0x00
'
' Server Notification
' NB : Username
' 1B : Player Number (which player number dropped)
'
' 0x15 = Ready to Play Signal
' Client Request:
' NB : Empty String [00]
' *Send this when your game is ready to start.*
'
' Server Notification:
' NB : Empty String [00]
' *Receive this when All Players are ready to start.*
'
' 0x16 = Connection Rejected
' Server Notification
' NB : Username
' 2B : UserID
' NB : Message
'
' 0x17 = Server Information Message
' Server Notification:
' NB : "Server\0"
' NB : Message
```

# flow

```
•••••••••••••••••••••••••••••
• Kaillera Network Protocol •
•••••••••••••••••••••••••••••
Kaillera network protocol analysis started originally by Anthem and Okai project.
Completed by SupraFast with help from Moosehead.
Master server format analysis by SupraFast.
'--------------------------------------------------------------------------------
'Protocol
'--------------------------------------------------------------------------------
'//Logging in State
'Client: HELLO0.83
'Server: Port notify HELLOD00D#\0 [# = new port number. eg. HELLOD00D7159]
' or TOO\0 [if server is full]
'Client: User Login Information [0x03]
'Server: Server to Client ACK [0x05]
'Client: Client to Server ACK [0x06]
'Server: Server to Client ACK [0x05]
'Client: Client to Server ACK [0x06]
'Server: Server to Client ACK [0x05]
'Client: Client to Server ACK [0x06]
'Server: Server to Client ACK [0x05]
'Client: Client to Server ACK [0x06]
'***ACK's calculate one's ping. Generally 4 are sent, but not always. Clients _
' respond to Server ACK's***
'Server: Server Status [0x04]
'Server: User Joined [0x02]
'Server: Server Information Message [0x17]
'
'//Global Chat State
'Client: Global Chat Request [0x07]
'Server: Global Chat Notification [0x07]
'
'//Game Chat State
'Client: Game Chat Request [0x08]
'Server: Global Chat Notification [0x08]
'
'//Create Game State
'Client: Create Game Request [0x0A]
'Server: Create Game Notification [0x0A]
'Server: Update Game Status [0x0A]
'Server: Player Information [0x0D]
'Server: Join Game Notification [0x0C]
'Server: Server Information Message [0x17]
'
'//Join Game State
'Client: Join Game Request [0x0C]
'Server: Update Game Status [0x0E]
'Server: Player Information [0x0D]
'Server: Join Game Notification [0x0C]
'
'//Quit Game State
'Client: Quit Game Request [0x0B]
'Server: Update Game Status [0x0E]
'Server: Quit Game Notification [0x0B]
'
'//Close Game State
'Client: Quit Game Request [0x0B]
'Server: Close Game Notification [0x10]
'Server: Quit Game Notification [0x0B]
'
'//Start Game State
'Client: Start Game Request [0x11]
'Server: Update Game Status [0x0E]
'Server: Start Game Notification [0x11]
'Client: *Netsync Mode* Wait for all players to give: Ready to Play Signal [0x15]
'Server: Update Game Status [0x0E]
'Server: *Playing Mode* All players are ready to start: Ready to Play Signal
Notification [0x15]
'Client(s): Exchange Data. Game Data Send [0x12] or Game Cache Send [0x13]
'Server: Sends data accordingly. Game Data Notify [0x12] or Game Cache Notify
[0x13]
'
'//Drop Game State
'Client: Drop Game Request [0x14]
'Server: Update Game Status [0x0E]
'Server: Drop Game Notification [0x14]
'
'//Kick Player State
'Client: Kick Request [0x0F]
'Server: Quit Game Notification [0x0B]
'Server: Update Game Status [0x0E]
'
'//User Quit State
'Client: User Quit Request [0x01]
'Server: User Quit Notification [0x01]
```

