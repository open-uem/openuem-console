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
			Usage:   "the path to your OCSP Console private key file in PEM format",
			EnvVars: []string{"CONSOLE_KEY_FILENAME"},
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
			Value:   "1323",
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
			Usage:    "the DNS domain used to contact with agents",
			EnvVars:  []string{"DOMAIN"},
			Required: true,
		},
	}
}
