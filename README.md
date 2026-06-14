# tumblr

Browse Tumblr blogs, posts, and tags from the command line.

`tumblr` is a single pure-Go binary. No API key required.

## Install

```bash
go install github.com/tamnd/tumblr-cli/cmd/tumblr@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/tumblr-cli/releases), or run the container image:

```bash
docker run --rm ghcr.io/tamnd/tumblr:latest --help
```

## Usage

```bash
# List posts tagged "photography"
tumblr tag photography

# List posts from a specific blog
tumblr posts staff

# Show blog metadata
tumblr info staff

# JSON output
tumblr tag art -o json -n 10

# Table output
tumblr posts nasa -o table
```

## Commands

| Command | Description |
|---------|-------------|
| `tag <tag>` | List posts tagged with a given tag |
| `posts <blog>` | List posts from a blog |
| `info <blog>` | Show blog metadata |
| `version` | Show version information |

## Global flags

```
-o, --output string    output format: table|json|jsonl|csv|tsv|url|raw (default "auto")
-n, --limit int        limit number of records (default 20)
    --fields strings   comma-separated columns to include
    --no-header        omit header row
    --template string  Go text/template per record
    --timeout duration per-request timeout (default 30s)
    --delay duration   minimum spacing between requests
    --retries int      retry attempts on 429/5xx (default 3)
```

## License

Apache-2.0. See [LICENSE](LICENSE).
