# ATlas

[Atproto](https://atproto.com/)-backed clone of [meetinghouse.cc](meetinghouse.cc/).

![Screenshot of the ATlas application](docs/screenshot.png)


## Build & Run
_If you don't use [mise](https://mise.jdx.dev/), you can just run the commands inside of `mise.toml` individually._

Install dependencies:
```bash
mise deps
```

Run locally:
```bash
mise dev 
```

Seeding the local DB with pins can be done by running `go run generate_mock_pins.go` in the `scripts/` directory.

To cut a build:
```bash
mise build
```