module github.com/xhebox/chrootd

go 1.13

replace github.com/xhebox/chrootd/commands => ./commands

require (
	github.com/sevlyar/go-daemon v0.1.5
	golang.org/x/sys v0.0.0-20200217220822-9197077df867 // indirect
)
