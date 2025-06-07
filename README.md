
# Chat-App Assignment

#### For Demo Video Link: [click here](https://drive.google.com/file/d/1-oA630KvQcyz7j_tibO6LIVgW94PTzmc/view?usp=sharing)

As a part of this assignment, I have created three microservices:

1. **authenticationService**
2. **websocketService**
3. **persistenceService**

## 🛠 Prerequisites

Ensure the following are installed and properly configured on your system:

- Git
- Docker
- Postman
- Golang (with appropriate `GOPATH` and `GOROOT` set)

## To run in one automatically using single command:
1. Clone the repo using command: `https://github.com/vickypl/chat-app.git`
2. Nevigate to cloned directory: `cd chap-app`
3. Execute command: `docker-compose -d` in your terminal.
(This will automatically setup all required containers)


## Usage:
Import following API curls in your postman:-
1. Token generator API postman curl [here](#collection)

2. The Chat API uses WebSocket connections for real-time communication. Since this is a WebSocket-based JSON API, you cannot share it directly as Postman's standard HTTP request functionality. You can follow websocketService ([here](#websocketService)) from service discription section below to use the same in your local postman.

## Enviroment setup - manual setup
***PostgressSQL setup***:- \
Run following commands on your terminal
to setup postgressDB on your local.
1. `docker run --name pg1 -e POSTGRES_PASSWORD=root -d postgres`
2. `docker exec -it pg1 psql -U postgres`
3. `create database chatapp`
4. `CREATE USER testuser1 WITH PASSWORD 'root';`
5. `GRANT ALL PRIVILEGES ON DATABASE chatapp TO testuser1;`
6. `GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO testuser1;`
7. `GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO testuser1;`
8. `GRANT ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public TO testuser1;`
\

***Kafka Setup***:- \
Run following commands on your terminal
to setup kafka on your local.
1. `docker run -d -p 2181:2181 -p 443:2008 -p 2008:2008 -p 2009:2009 --env ADVERTISED_LISTENERS=PLAINTEXT://kafka:443,INTERNAL://localhost:2009 --env LISTENERS=PLAINTEXT://0.0.0.0:2008,INTERNAL://0.0.0.0:2009 --env SECURITY_PROTOCOL_MAP=PLAINTEXT:PLAINTEXT,INTERNAL:PLAINTEXT --env INTER_BROKER=INTERNAL --env KAFKA_CREATE_TOPICS="test:36:1,krisgeus:12:1:compact" --name kafka-server krisgeus/docker-kafka\`
2. `docker exec -it kafka-server bash`
3. `cd opt/kafka_2.12-2.3.0/bin/`
4. `./kafka-topics.sh --create --bootstrap-server localhost:2009 --replication-factor 1 --partitions 1 --topic messages`
5. `./kafka-topics.sh --create --bootstrap-server localhost:2009 --replication-factor 1 --partitions 1 --topic dead-messages`
6. `exit`


### Services Discription:-
1. **authenticationService**: This service is responsible for generating JWT auth token for given username and password as we are doing stateless authentication. It contains a `configs/.env` directory containing secret used for generating authtoken.

```
Note: For simplicity and due time constraint i have hard coded some dummy user inside the code authentication service itself, for actuall use we can use database or some other persistent storage.

This are the following dummy usernames and there password i have hardcoded for testing purpose, you can use this to generate tokens:
	user1=pwd1
	user2=pwd2
	user3=pwd3
	user4=pwd4
	user5=pwd5
```

`Note: all below steps applicable post cloning the repository in your go working directory using the command: https://github.com/vickypl/chat-app.git` 


#### steps to run the service:-
1. Open terminal(say T1) at path chat-app directory.
2. Nevigate inside authenticationService directory using 
   command `cd authenticationService`.
3. Run the app using: `go run main.go` (This will run the app on port 8000)

### Use the following curl from your postman to check if token is generated successffuly or not
Request:-
<a id="collection"></a>
```
curl --location --request GET 'localhost:8000/login' \
--header 'Content-Type: application/json' \
--data '{
    "username": "user1",
    "password": "pwd1"
}'
```
Response(example):-
```
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDkwNTUyMTAsInVzZXJuYW1lIjoidXNlcjEifQ.I_mWqo_SNP6yfYZDnJrM6Yo3wC3RIqcF_9v_wbvBLuI"
}
```
---
<a id="websocketService"></a>
2. **websocketService**: This service is reponsible for handling websocket connections and user realtime chats, It contains a `configs/.env` directory containing various enviroment varibles required to run this service. Apart from handling messages exchanges this service publishes the messages on a kafka topic `messages` to store the messages persistently.
#### steps to run the service:-
1. Open another terminal(say T2) at path chat-app directory.
2. Nevigate inside websocketService directory using 
   command `cd websocketService`.
3. Run the app using: `go run main.go` (This will run the app on port 8080)

### Use the following curl from your postman to check if you are able to connect with the chat server
Request:- (This is not curl request, you menually have to use it in postman)
```
Url: localhost:8080/ws/send_uuid
Example: localhost:8080/ws/123e4567-e89b-12d3-a456-426614174000
Request Body:
{
  "sender_id": "123e4567-e89b-12d3-a456-426614174000",
  "recipient_id": "789e0123-e45b-67d8-a901-324534567890",
  "content": "Hello! How are you doing today?"
}

Header:
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDkwNTUyMTAsInVzZXJuYW1lIjoidXNlcjEifQ.I_mWqo_SNP6yfYZDnJrM6Yo3wC3RIqcF_9v_wbvBLuI

Note: In header you have to pass the jwt token that is generated using the authenticationService with Bearer as prefix as given in example
```
Response(example):-
```
This will start the connection with the socket server. You can try exchanging messages post then by similary connecting with another recipient id in other tab.
```
---
3. **persistenceService**: This service is responsible for asynchronsly storing messages exachanged by user into persistent storage(here postgresSQL DB). It contains a `configs/.env` directory containing various enviroment varibles required to run this service along with there discription.

Working: This service will consume messages post by websocketService on `messages` topic, and it will store those messages inside postgress sql DB. In case it fails to store the messages(might be due to db error or network error) it will re-publish the message to messages topic and reprocess the message till MAX_RETRY_VALUE given in .env file.

For example: if we have given MAX_RETRY_VALUE as 3, it will re-publish the message on `messages` topic max 3 times. If still not successfully store the message into postgressSQL it will publish that message to another topic `dead-messages`. In future we can have something like dead-messages processor cron job which can republish this dead messages from the que to store in DB.

#### steps to run the service:-
1. Open another terminal(say T3) at path chat-app directory.
2. Nevigate inside persistenceService directory using 
   command `cd persistenceService`.
3. Run the app using: `go run main.go` (This will start the consumer)

---
#### End to end flow for testing:-
1. Have docker, git, postman in your local machine.
2. Clone the chat app repostry
3. Setup env `postgressSQL, kafka` using the above steps. (using the same commands will be more quicker and useful since i have given that as per .env variables i used for testing)
4. Run authenticationService in Terminal1 (following steps given in discription)
5. Run websocketService in Terminal2 (following steps given in discription)
6. Run persistenceService in Terminal3 (following steps given in discription)
7. Generate auth token using dummy username and passwords provided.
8. Make web socket connection as discribed in websocketService via postman tab1 and tab2 using the generated auth token.
9. Once you'll start sending message, you can check in your postgress DB's messages table. You will be able to see the stored messages.