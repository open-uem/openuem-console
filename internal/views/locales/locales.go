package locales

import "embed"

//go:embed de.yaml
//go:embed en.yaml
//go:embed es.yaml
//go:embed ca.yaml
//go:embed fr.yaml

var Content embed.FS
