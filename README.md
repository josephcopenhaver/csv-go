# csv-go

[![Go Report Card](https://goreportcard.com/badge/github.com/josephcopenhaver/csv-go)](https://goreportcard.com/report/github.com/josephcopenhaver/csv-go/v2)
![tests](https://github.com/josephcopenhaver/csv-go/actions/workflows/tests.yaml/badge.svg)
![code-coverage](https://img.shields.io/badge/code_coverage-100%25-rgb%2852%2C208%2C88%29)

## Why does this exist?
I am tired of rewriting this over and over to cover edge cases where other language standard csv implementations have assertions on the format and formatting I cannot guarantee are valid for a given file and how it was constructed. I've written variations that cover far fewer concerns over the years, and I figured I'll make a superset of one that does everything I need and then experiment with making it as efficient as maintainability will allow. Feel free to use however you may wish.

[CHANGELOG](./docs/version/v2/CHANGELOG.md)

---

[![Go Documentation](https://godocs.io/github.com/josephcopenhaver/csv-go/v2?status.svg)](https://godocs.io/github.com/josephcopenhaver/csv-go/v2)
