# ugfetch

`ugfetch` downloads Ultimate Guitar tabs by id or song URL, writes them into an artist folder, and can transpose them to a target key.

## Usage

```sh
ugfetch <ug-id|song-url> [--key KEY] [--markup] [--output-dir PATH]
```

## Examples

```sh
ugfetch 123456
ugfetch https://www.ultimate-guitar.com/tab/example/song_123456 --key G
ugfetch 123456 --markup --output-dir ./tabs
```

## Build

```sh
go build ./...
```

