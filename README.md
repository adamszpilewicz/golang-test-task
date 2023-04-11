docker-compose up -d

- POST endpoint: `localhost:8088/message` with body:
```json
{
    "sender": "Adam",
    "receiver": "Rena",
    "message": "Hello, Rena 0000!"
}
```

- GET endpoint: `localhost:8089/message/list?sender=Adam&receiver=Rena`

