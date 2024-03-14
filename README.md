# Webhooker

Webhooker is an example implementation of webhook service. It allows for posting
a message to the webhook. The list of all posted message can then be displayed.

## Installation

```bash
go build webhooker.go
```

## Usage
To create a new service token:
```bash
webhooker createToken --dburi <mongodb_connection_string>
```

To start the webhook service, execute:
```bash
webhooker listen --certfile <certificate.crt> --keyfile <certificate.key> --dburi <mongodb_connection_string> --listen-address [ip_address]<:port>
```

To post a message to the webhook service:
```bash
curl -k -X POST -H 'Content-Type: application/json' -d '{"token":"<token>","channel":"<channel_not_used_yet>","message":"<message_to_be_posted>"}' https://<ip_address>:<port>/webhook
```

To display already posted messages, visit:
```bash
curl -khttps://<ip_address>:<port>/wall
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

## License

GPL 3.0
