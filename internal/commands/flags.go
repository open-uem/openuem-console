package commands

import "github.com/urfave/cli/v2"

func StartConsoleFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "cacert",
			Value:   "certificates/ca.cer",
			Usage:   "the path to OpenUEM's CA certificate file in PEM format",
			EnvVars: []string{"CA_CRT_FILENAME"},
		},
		&cli.StringFlag{
			Name:    "cert",
			Value:   "certificates/console.cer",
			Usage:   "the path to OpenUEM's Console certificate file in PEM format",
			EnvVars: []string{"CONSOLE_CERT_FILENAME"},
		},
		&cli.StringFlag{
			Name:    "key",
			Value:   "certificates/console.key",
			Usage:   "the path to your OpenUEM's Console private key file in PEM format",
			EnvVars: []string{"CONSOLE_KEY_FILENAME"},
		},
		&cli.StringFlag{
			Name:    "sftpkey",
			Value:   "certificates/sftp.key",
			Usage:   "the path to your SFTP certificate private key file in PEM format",
			EnvVars: []string{"SFTP_KEY_FILENAME"},
		},
		&cli.StringFlag{
			Name:     "nats-servers",
			Usage:    "comma-separated list of NATS servers urls e.g (tls://localhost:4433)",
			EnvVars:  []string{"NATS_SERVERS"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "dburl",
			Usage:    "the Postgres database connection url e.g (postgres://user:password@host:5432/openuem)",
			EnvVars:  []string{"DATABASE_URL"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "jwt-key",
			Usage:    "a string signed to use JWT tokens used in email address confirmation",
			EnvVars:  []string{"JWT_KEY"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "server-name",
			Usage:   "the server name like example.com or localhost",
			EnvVars: []string{"SERVER_NAME"},
			Value:   "localhost",
		},
		&cli.StringFlag{
			Name:    "console-port",
			Usage:   "the TCP port used by the console server",
			EnvVars: []string{"CONSOLE_PORT"},
			Value:   "1323",
		},
		&cli.StringFlag{
			Name:    "auth-port",
			Usage:   "the TCP port used by the authentication server",
			EnvVars: []string{"AUTH_PORT"},
			Value:   "1324",
		},
		&cli.StringFlag{
			Name:     "domain",
			Usage:    "the DNS domain used to contact agents",
			EnvVars:  []string{"DOMAIN"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "org-name",
			Usage:    "the name of your organization as it appears in digital certificates",
			EnvVars:  []string{"ORGNAME"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "org-province",
			Usage:   "the province of your organization as it appears in digital certificates",
			EnvVars: []string{"ORGPROVINCE"},
		},
		&cli.StringFlag{
			Name:    "org-locality",
			Usage:   "the locality of your organization as it appears in digital certificates",
			EnvVars: []string{"ORGLOCALITY"},
		},
		&cli.StringFlag{
			Name:    "org-address",
			Usage:   "the address of your organization as it appears in digital certificates",
			EnvVars: []string{"ORGADDRESS"},
		},
		&cli.StringFlag{
			Name:    "country",
			Usage:   "the country of your organization as it appears in digital certificates",
			EnvVars: []string{"COUNTRY"},
		},
		&cli.StringFlag{
			Name:    "reverse-proxy-auth-port",
			Usage:   "if the console runs behind a reverse proxy, we need to set the authentication server port",
			EnvVars: []string{"REVERSE_PROXY_AUTH_PORT"},
		},
		&cli.StringFlag{
			Name:    "reverse-proxy-server",
			Usage:   "if the console runs behind a reverse proxy, we need to know the console domain",
			EnvVars: []string{"REVERSE_PROXY_SERVER"},
		},
		&cli.BoolFlag{
			Name:    "re-enable-certificates-auth",
			Usage:   "if you disabled the use of certificates to log in and cannot use OIDC you can re-enable it again",
			EnvVars: []string{"RE_ENABLE_CERTIFICATES_AUTH"},
			Value:   false,
		},
		&cli.BoolFlag{
			Name:    "re-enable-passwd-auth",
			Usage:   "if you disabled the use of passwords to log in you can re-enable it again",
			EnvVars: []string{"RE_ENABLE_PASSWD_AUTH"},
			Value:   false,
		},
		&cli.BoolFlag{
			Name:    "reset-openuem-user",
			Usage:   "this will set a new password for the openuem account and disable 2FA",
			EnvVars: []string{"RESET_OPENUEM_USER"},
			Value:   false,
		},
	}
}
