# rss2dayone

Reads an RSS (or Atom) feed. Adds every post to your [Day One](https://dayoneapp.com) journal.

Based on a [python version](https://github.com/benjamineskola/scripts/blob/455cc5026a438c6156249c5cf01432a16dc86b29/rss2dayone.py) that got a little unwieldy.

## Installation

```sh
$ go install github.com/benjamineskola/rss2dayone
```

Requires Go, because I haven't tried building binaries to distribute.

Requires the Day One app to be installed, and the command-line tool `dayone2` to be in your path. This means it requires macOS, probably, because that's the only platform Day One will run on that you could also run a Go program on.

## Usage

```sh
rss2dayone
```

## Configuration

The configuration file lives in `$XDG_CONFIG_HOME/rss2dayone.toml`, which is by default `$HOME/.config/rss2dayone.toml`. The format is as follows:

```toml
[[feed]]
url = "http://example.com/feed.xml"
journal = "Journal"
tags = ["example"]

[[feed]]
url = "https://anotherexample.net/news.rss"
journal = "Another Journal"
# tags are optional
```

## Features in progress/done

- [x] Attachment support
- [ ] Concurrency (though it may be that Day One's database is the bottleneck anyway)
- Special cases for particular feeds:
  - [x] [Letterboxd](https://letterboxd.com)
  - [ ] Mastodon
  - [ ] an extensible manner rather than hardcoding every feed
