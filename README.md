docker-compose up -d

- POST endpoint: localhost:8088/message with body:
{
    "sender": "Adam",
    "receiver": "Rena",
    "message": "Hello, Rena 0000!"
}

- GET endpoint: localhost:8088/message?sender=Adam&receiver=Rena

