module gopkg.in/kainz/cellmodemd.v0

go 1.18

require (
	github.com/google/renameio v1.0.1
	github.com/maltegrosse/go-modemmanager v0.1.1-0.20211120100021-a70df96fe495
)

replace github.com/maltegrosse/go-modemmanager => ../go-modemmanager/

replace github.com/godbus/dbus/v5 => github.com/godbus/dbus/v5 v5.0.6

require github.com/godbus/dbus/v5 v5.0.3 // indirect
