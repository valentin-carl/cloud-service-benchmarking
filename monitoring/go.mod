module monitoring

go 1.21

require (
	benchmark/lib v0.0.0-00010101000000-000000000000
	github.com/mackerelio/go-osstat v0.2.4
)

require (
	github.com/VividCortex/multitick v1.0.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
)

replace benchmark/lib => ../lib
