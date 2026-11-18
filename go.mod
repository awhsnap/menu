module menuapp

go 1.24.7

require (
	github.com/mattn/go-sqlite3 v1.14.53
	golang.org/x/crypto v0.23.0
)

replace golang.org/x/crypto => github.com/golang/crypto v0.23.0
