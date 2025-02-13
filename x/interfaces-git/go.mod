module github.com/stevegt/grid-poc/x/interfaces-git

go 1.22.1

replace github.com/stevegt/grid-poc/x/cbor-codec => /home/stevegt/lab/grid-poc/x/cbor-codec

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/fxamacker/cbor/v2 v2.7.0
	github.com/stevegt/goadapt v0.7.0
	github.com/stevegt/grid-poc/x/cbor-codec v0.0.0-00010101000000-000000000000
)

require github.com/x448/float16 v0.8.4 // indirect
