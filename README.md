<p align="center"><img src="https://raw.githubusercontent.com/go-ruby-minitest/brand/main/social/go-ruby-minitest.png" alt="go-ruby-minitest" width="720"></p>

# docs — go-ruby-minitest

The documentation site for [go-ruby-minitest](https://github.com/go-ruby-minitest),
built with [MkDocs Material](https://squidfunk.github.io/mkdocs-material/) and
versioned with [mike](https://github.com/jimporter/mike). Served at
[go-ruby-minitest.github.io/docs/](https://go-ruby-minitest.github.io/docs/).

## Build locally

```sh
pip install -r requirements.txt
mkdocs serve
```

## Benchmarks

The cross-runtime performance harness behind
[`docs/performance.md`](docs/performance.md) lives in
[`benchmarks/`](benchmarks/): a self-contained Go driver (`go/`, pins the
published library by pseudo-version — no `replace`), the equivalent
`ruby/minitest.rb` workload, and `run.sh`. See
[`benchmarks/README.md`](benchmarks/README.md).

## Deploy

Pushing to `main` runs `.github/workflows/docs.yml`, which `mike deploy`s the
versioned site to the `gh-pages` branch (served by GitHub Pages).

## License

BSD-3-Clause — see [LICENSE](LICENSE). Copyright the go-ruby-minitest/docs authors.
