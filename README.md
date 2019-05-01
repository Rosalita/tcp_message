# tcp_message
Server / client messaging over TCP

## To Build Hub

- change into the hub directory
- build the hub binary

`cd hub/ && go build`

## To Build Client

- change into the client directory
- build the client binary

`cd client/ && go build`


## Run Hub first

`./hub`

## Then run clients

Start client and immediately send an identity message

`./client -cmd identity`

Start client and immediately send a list message

`./client -cmd list`

Start client and immediately send a relay message

`./client -cmd relay -m hi -to 1,2`


## Commands

# Identity
After sending an identity message, the hub will respond with the user_id of the connected user.

# List
After sending a list message, the hub will answer with the list of all connected client user_id’s (excluding the requesting client)

# Relay
After sending a relay message, the `-m message` is relayed `-to 1,2` the recipient client ids.

Maximum number of recipients for relay messages is 255


# Mockhub

Mockhub is a minimal hub package used by the client's unit tests