module gopkg.in/kainz/cellmodemd.v0

go 1.18

require github.com/google/renameio v1.0.1

replace github.com/maltegrosse/go-modemmanager => github.com/kainz/go-modemmanager v0.1.0-modemmanager-1.10

replace github.com/godbus/dbus/v5 => github.com/godbus/dbus/v5 v5.0.6

require (
	github.com/godbus/dbus/v5 v5.0.3 // indirect
	github.com/maltegrosse/go-modemmanager v0.1.1-0.20221114040231-8420f1b68d04 // indirect
)
