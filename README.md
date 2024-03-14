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

## Contributing

Pull requests are welcome. For major changes, please open an issue first
to discuss what you would like to change.

## License

GPL 3.0
