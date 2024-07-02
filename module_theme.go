package neomega_backbone

type Theme interface {
	// variate can be a variate type of a theme
	// like "dark", "light", "plain", etc.
	// or a key of a specific object
	// like player name such as "2401PT"
	// chain call is supported
	// like omega.GetTheme().GetVariate("2401PT").GetVariate("dark"), which means "2401PT" in dark mode
	// while omega.GetTheme() provides the default theme
	GetVariate(style string) Theme
	// set aesthetics of control how it looks like
	// e.g. SetAesthetics("theme_color", "blue") will set theme color to blue
	// e.g. SetAesthetics("error_color", "red") will set error color to red
	// e.g. SetAesthetics("hint.unknown_selection.text", "sorry i don't understand your choice")
	// will set the hint text of unknown selection to "sorry i don't understand your choice"
	// e.g. SetAesthetics("hint.unknown_selection.color", "red")
	// will set the hint color of unknown selection to red, and if you don't specify the color, it will use the default "error" color
	SetAesthetics(key string, value string)
	// render a specific object
	// e.g. Render("hint.unknown_selection",[selection],[available_selections])
	// will generate a hint text for unknown selection with the style basing on the current theme
	Render(key string, args ...any) string
}

type ExtendTheme interface {
	GetTheme() Theme
	SetTheme(theme Theme)
	GetVariate(style string) ExtendTheme
	SetAesthetics(key string, value string)
	Render(key string, args ...any) string
	// and many utility functions like:
	GenStringListHintResolver(available []string) (string, func(params []string) (selection int, cancel bool, err error))
	GenStringListHintResolverWithIndex(available []string) (string, func(params []string) (selection int, cancel bool, err error))
	GenIntRangeResolver(min int, max int) (string, func(params []string) (selection int, cancel bool, err error))
	GenYesNoResolver() (string, func(params []string) (bool, error))
	// and many other functions
}
