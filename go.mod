module github.com/Gleipnir-Technology/fieldseeker-sync

go 1.24

toolchain go1.24.3

require github.com/Gleipnir-Technology/arcgis-go v0.0.2 // explicit

require (
	github.com/Gleipnir-Technology/fieldseeker-sync/shared v0.0.0
	github.com/alexedwards/scs/v2 v2.8.0
	github.com/fsnotify/fsnotify v1.8.0
	github.com/gdamore/tcell/v2 v2.8.1
	github.com/georgysavva/scany/v2 v2.1.4
	github.com/go-chi/chi/v5 v5.2.1
	github.com/go-chi/render v1.0.3
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.4
	github.com/pressly/goose/v3 v3.24.3
	github.com/spf13/viper v1.20.1
	golang.org/x/crypto v0.38.0
)

require (
	github.com/MakeNowJust/heredoc v1.0.0 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/alessio/shellescape v1.4.1 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/briandowns/spinner v1.23.2 // indirect
	github.com/charmbracelet/colorprofile v0.2.3-0.20250311203215-f60798e515dc // indirect
	github.com/charmbracelet/lipgloss v1.1.0 // indirect
	github.com/charmbracelet/x/ansi v0.8.0 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.13-0.20250311204145-2c3ea96c31dd // indirect
	github.com/charmbracelet/x/term v0.2.1 // indirect
	github.com/creack/pty v1.1.24 // indirect
	github.com/evilmartians/lefthook v1.11.13 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/gdamore/encoding v1.0.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/goccy/go-yaml v1.17.1 // indirect
	github.com/gotnospirit/makeplural v0.0.0-20180622080156-a5f48d94d976 // indirect
	github.com/gotnospirit/messageformat v0.0.0-20221001023931-dfe49f1eb092 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kaptinlin/go-i18n v0.1.3 // indirect
	github.com/kaptinlin/jsonschema v0.2.3 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/parsers/json v1.0.0 // indirect
	github.com/knadh/koanf/parsers/toml/v2 v2.2.0 // indirect
	github.com/knadh/koanf/parsers/yaml v1.0.0 // indirect
	github.com/knadh/koanf/providers/fs v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mattn/go-tty v0.0.7 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/sagikazarmark/locafero v0.7.0 // indirect
	github.com/schollz/progressbar/v3 v3.18.0 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.14.0 // indirect
	github.com/spf13/cast v1.7.1 // indirect
	github.com/spf13/cobra v1.9.1 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/xo/terminfo v0.0.0-20220910002029-abceb7e1c41e // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.14.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/term v0.32.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/Gleipnir-Technology/fieldseeker-sync/html => ./html

replace github.com/Gleipnir-Technology/fieldseeker-sync/shared => ./shared

tool github.com/evilmartians/lefthook
