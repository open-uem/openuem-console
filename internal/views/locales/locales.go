package locales

import "embed"

//go:embed de.yaml
//go:embed en.yaml
//go:embed es.yaml
//go:embed ca.yaml
//go:embed fr.yaml
//go:embed no.yaml
//go:embed pt.yaml

var Content embed.FS
