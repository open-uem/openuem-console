module github.com/doncicuto/openuem-console

go 1.23.1

require (
	entgo.io/ent v0.14.1
	github.com/a-h/templ v0.2.793
	github.com/alexedwards/scs/pgxstore v0.0.0-20240316134038-7e11d57e8885
	github.com/alexedwards/scs/v2 v2.8.0
	github.com/biter777/countries v1.7.5
	github.com/canidam/echo-scs-session v1.0.0
	github.com/danieljoos/wincred v1.2.2
	github.com/doncicuto/openuem_ent v0.0.0-00010101000000-000000000000
	github.com/doncicuto/openuem_nats v0.0.0-00010101000000-000000000000
	github.com/doncicuto/openuem_utils v0.0.0-00010101000000-000000000000
	github.com/go-co-op/gocron/v2 v2.12.1
	github.com/go-echarts/go-echarts/v2 v2.4.1
	github.com/go-playground/form/v4 v4.2.1
	github.com/go-playground/validator/v10 v10.22.1
	github.com/gohugoio/locales v0.14.0
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/gomarkdown/markdown v0.0.0-20240930133441-72d49d9543d8
	github.com/invopop/ctxi18n v0.8.1
	github.com/jackc/pgx/v5 v5.6.0
	github.com/labstack/echo/v4 v4.12.0
	github.com/linde12/gowol v0.0.0-20180926075039-797e4d01634c
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/mssola/useragent v1.0.0
	github.com/nats-io/nats.go v1.37.0
	github.com/pablor21/echo-etag/v4 v4.0.3
	github.com/pkg/sftp v1.13.6
	github.com/urfave/cli/v2 v2.27.5
	github.com/wneessen/go-mail v0.5.1
	golang.org/x/crypto v0.28.0
	golang.org/x/mod v0.20.0
	golang.org/x/sys v0.26.0
	gopkg.in/ini.v1 v1.67.0
	software.sslmate.com/src/go-pkcs12 v0.5.0
)

require (
	ariga.io/atlas v0.19.1-0.20240203083654-5948b60a8e43 // indirect
	github.com/agext/levenshtein v1.2.1 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.5 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/go-openapi/inflect v0.19.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/hashicorp/hcl/v2 v2.13.0 // indirect
	github.com/invopop/yaml v0.3.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/klauspost/compress v1.17.2 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-wordwrap v0.0.0-20150314170334-ad45545899c7 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	github.com/zclconf/go-cty v1.8.0 // indirect
	golang.org/x/exp v0.0.0-20240613232115-7f521ea00fb8 // indirect
	golang.org/x/net v0.28.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/doncicuto/openuem_ent => ./internal/models/ent

replace github.com/doncicuto/openuem_nats => ./internal/controllers/nats

replace github.com/doncicuto/openuem_utils => ./internal/controllers/utils
