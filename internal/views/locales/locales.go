package locales

import "embed"

//go:embed de.yaml
//go:embed en.yaml
//go:embed es.yaml

var Content embed.FS
