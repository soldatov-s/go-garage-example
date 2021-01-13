package cmd

const defAppFieldValue = "undefined"

var (
	// nolint : global var for application info
	version = defAppFieldValue
	// nolint : global var for application info
	builded = defAppFieldValue
	// nolint : global var for application info
	hash = defAppFieldValue
	// nolint : global var for application info
	appName = defAppFieldValue
	// nolint : global var for application info
	description = ""

	// nolint : global var for application info
	appFullVersion = version + ", builded: " + builded + ", hash: " + hash
)
