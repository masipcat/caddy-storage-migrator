# Caddy Storage Migrator

[![Pipeline](https://gitlab.masip.cat/masipcat/caddy-storage-migrator/badges/master/pipeline.svg)](https://gitlab.masip.cat/masipcat/caddy-storage-migrator/-/commits/master) [![Coverage](https://gitlab.masip.cat/masipcat/caddy-storage-migrator/badges/master/coverage.svg)](https://gitlab.masip.cat/masipcat/caddy-storage-migrator/-/commits/master)

This is a simple command line application to import/export storage data from Caddy v2 (or Certmagic).

Right now supports the following modules:

- `redis` ([caddy-tlsredis](https://github.com/gamalan/caddy-tlsredis))

## Build

First we need to build the command:

```sh
cd cmd/migrator
go build
```

## Usage

Now we can **import** existing data with the following command:

```sh
./migrator import /path/to/existing/caddypath <module-name>
```

...or **export** the redis data to the filesystem:

```sh
./migrator export <module-name> ./any/folder
```

Optionally the command accepts the flag `-config path/to/file.json`.
The JSON file should contain at least the storage configuration:

```json
{
    "storage": {
        "..."
    }
}
```
