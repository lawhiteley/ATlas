# ATlas

[Atproto](https://atproto.com/)-backed clone of [meetinghouse.cc](meetinghouse.cc/).

![Screenshot of the ATlas application](docs/screenshot.png)


## Build & Run
_If you don't use [mise](https://mise.jdx.dev/), you can just run the commands inside of `mise.toml` individually._

### Install dependencies:
```bash
mise deps
```

### Run locally:
```bash
mise dev 
```

Seeding the local DB with pins can be done by running `go run generate_mock_pins.go` in the `scripts/` directory.

### Cut a build:
```bash
mise build
```

## Deployment

The service should run locally without any of the below due to the [development client](https://atproto.com/specs/oauth#localhost-client-development) exception for `localhost`. If you wish to deploy this however, you'll need to enable the [confidential client](https://atproto.com/specs/oauth#localhost-client-development) by setting the following values: 


* `SESSION_SECRET`: Any ol' secret. Something like `openssl rand -hex 32`
* `CLIENT_HOSTNAME`: Wherever the service is running (e.g. https://atlas.whiteley.io)
* `CLIENT_SECRET_KEY`: EC private key in Multibase format. Can be generated with `goat key generate -t P-256`
