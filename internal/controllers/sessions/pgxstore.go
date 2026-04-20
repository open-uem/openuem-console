package sessions

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/open-uem/utils"
)

// This code is cherry-picked from https://github.com/alexedwards/scs/blob/209de6e426de9259665975ce16b91331d228f052/pgxstore/pgxstore.go#L11
// so we can support token encryption/decryption

// PostgresStore represents the session store.
type PostgresStore struct {
	pool                *pgxpool.Pool
	stopCleanup         chan bool
	tableName           string
	encryptionMasterKey string
}

type Config struct {
	// CleanUpInterval is the interval between each cleanup operation.
	// If set to 0, the cleanup operation is disabled.
	CleanUpInterval time.Duration

	// TableName is the name of the table where the session data will be stored.
	// If not set, it will default to "sessions".
	TableName string

	// EncryptionMasterKey is the key used to encrypt tokens
	EncryptionMasterKey string
}

// New returns a new PostgresStore instance, with a background cleanup goroutine
// that runs every 5 minutes to remove expired session data.
func NewPostgresStore(pool *pgxpool.Pool, encryptionMasterKey string) *PostgresStore {
	return NewWithConfig(pool, Config{
		CleanUpInterval:     5 * time.Minute,
		EncryptionMasterKey: encryptionMasterKey,
	})
}

// NewWithCleanupInterval returns a new PostgresStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewWithCleanupInterval(pool *pgxpool.Pool, cleanupInterval time.Duration) *PostgresStore {
	return NewWithConfig(pool, Config{
		CleanUpInterval: cleanupInterval,
	})
}

// NewWithConfig returns a new PostgresStore instance with the given configuration.
// If the TableName field is empty, it will be set to "sessions".
// If the CleanUpInterval field is 0, the cleanup goroutine will not be started.
func NewWithConfig(pool *pgxpool.Pool, config Config) *PostgresStore {
	if config.TableName == "" {
		config.TableName = "sessions"
	}

	p := &PostgresStore{pool: pool, tableName: config.TableName, encryptionMasterKey: config.EncryptionMasterKey}
	if config.CleanUpInterval > 0 {
		p.stopCleanup = make(chan bool)
		go p.startCleanup(config.CleanUpInterval)
	}
	return p
}

// FindCtx returns the data for a given session token from the PostgresStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (p *PostgresStore) FindCtx(ctx context.Context, token string) (b []byte, found bool, err error) {
	if p.encryptionMasterKey != "" {
		stmt := fmt.Sprintf("SELECT token, data FROM %s WHERE current_timestamp < expiry", p.tableName)
		rows, err := p.pool.Query(ctx, stmt)
		if err != nil {
			return nil, false, nil
		}
		defer rows.Close()
		for rows.Next() {
			var retrievedToken string
			err = rows.Scan(&retrievedToken, &b)
			if err != nil {
				return nil, false, nil
			}

			isAccessTokenEncrypted, err := utils.IsSensitiveFieldEncrypted(retrievedToken, p.encryptionMasterKey)
			if err != nil {
				return nil, false, nil
			}

			if isAccessTokenEncrypted {
				retrievedToken, err = utils.DecryptSensitiveField(retrievedToken, p.encryptionMasterKey)
				if err != nil {
					return nil, false, nil
				}

				if retrievedToken == token {
					return b, true, nil
				}
			}
		}
	} else {
		stmt := fmt.Sprintf("SELECT data FROM %s WHERE token = $1 AND current_timestamp < expiry", p.tableName)
		row := p.pool.QueryRow(ctx, stmt, token)
		err = row.Scan(&b)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		} else if err != nil {
			return nil, false, err
		}
		return b, true, nil
	}

	return nil, false, nil
}

// CommitCtx adds a session token and data to the PostgresStore instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (p *PostgresStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) (err error) {

	// encrypt the token in database if we have the encryption master key
	if p.encryptionMasterKey != "" {
		foundToken := ""
		stmt := fmt.Sprintf("SELECT token FROM %s WHERE current_timestamp < expiry", p.tableName)
		rows, err := p.pool.Query(ctx, stmt)
		if err != nil {
			return err
		}

		defer rows.Close()
		for rows.Next() {
			var retrievedToken string
			err = rows.Scan(&retrievedToken)
			if err != nil {
				return err
			}
			isAccessTokenEncrypted, err := utils.IsSensitiveFieldEncrypted(retrievedToken, p.encryptionMasterKey)
			if err != nil {
				return err
			}

			if isAccessTokenEncrypted {
				decryptedRetrievedToken, err := utils.DecryptSensitiveField(retrievedToken, p.encryptionMasterKey)
				if err != nil {
					return err
				}

				if decryptedRetrievedToken == token {
					foundToken = retrievedToken
				}
			}
		}

		if foundToken != "" {
			// token already exists encrypted, update
			token = foundToken
		} else {
			// encrypt token, this will be a new token
			token, err = utils.EncryptSensitiveField(token, p.encryptionMasterKey)
			if err != nil {
				return err
			}
		}
	}

	stmt := fmt.Sprintf("INSERT INTO %s (token, data, expiry) VALUES ($1, $2, $3) ON CONFLICT (token) DO UPDATE SET data = EXCLUDED.data, expiry = EXCLUDED.expiry", p.tableName)
	_, err = p.pool.Exec(ctx, stmt, token, b, expiry)
	return err
}

// DeleteCtx removes a session token and corresponding data from the PostgresStore
// instance.
func (p *PostgresStore) DeleteCtx(ctx context.Context, token string) (err error) {
	if p.encryptionMasterKey != "" {
		stmt := fmt.Sprintf("SELECT token FROM %s WHERE current_timestamp < expiry", p.tableName)
		rows, err := p.pool.Query(ctx, stmt)
		if err != nil {
			return err
		}

		defer rows.Close()
		for rows.Next() {
			var retrievedToken string
			err = rows.Scan(&retrievedToken)
			if err != nil {
				return err
			}
			isAccessTokenEncrypted, err := utils.IsSensitiveFieldEncrypted(retrievedToken, p.encryptionMasterKey)
			if err != nil {
				log.Println(err)

				return err
			}

			if isAccessTokenEncrypted {
				decryptedRetrievedToken, err := utils.DecryptSensitiveField(retrievedToken, p.encryptionMasterKey)
				if err != nil {
					return err
				}

				if decryptedRetrievedToken == token {
					stmt := fmt.Sprintf("DELETE FROM %s WHERE token = $1", p.tableName)
					_, err = p.pool.Exec(ctx, stmt, retrievedToken)
					return err
				}
			}
		}
	} else {
		stmt := fmt.Sprintf("DELETE FROM %s WHERE token = $1", p.tableName)
		_, err = p.pool.Exec(ctx, stmt, token)
		return err
	}

	return nil
}

// AllCtx returns a map containing the token and data for all active (i.e.
// not expired) sessions in the PostgresStore instance.
func (p *PostgresStore) AllCtx(ctx context.Context) (map[string][]byte, error) {
	stmt := fmt.Sprintf("SELECT token, data FROM %s WHERE current_timestamp < expiry", p.tableName)
	rows, err := p.pool.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make(map[string][]byte)

	for rows.Next() {
		var (
			token string
			data  []byte
		)

		err = rows.Scan(&token, &data)
		if err != nil {
			return nil, err
		}

		// decrypt token if key is set
		if p.encryptionMasterKey != "" {
			isTokenEncrypted, err := utils.IsSensitiveFieldEncrypted(token, p.encryptionMasterKey)
			if err != nil {
				return nil, err
			}
			if isTokenEncrypted {
				token, err = utils.DecryptSensitiveField(token, p.encryptionMasterKey)
				return nil, err
			}
		}

		sessions[token] = data
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (p *PostgresStore) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := p.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-p.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the PostgresStore
// instance. It's rare to terminate this; generally PostgresStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the PostgresStore is transient.
// An example is creating a new PostgresStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// PostgresStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (p *PostgresStore) StopCleanup() {
	if p.stopCleanup != nil {
		p.stopCleanup <- true
	}
}

func (p *PostgresStore) deleteExpired() error {
	stmt := fmt.Sprintf("DELETE FROM %s WHERE expiry < current_timestamp", p.tableName)
	_, err := p.pool.Exec(context.Background(), stmt)
	return err
}

// We have to add the plain Store methods here to be recognized a Store
// by the go compiler. Not using a separate type makes any errors caught
// only at runtime instead of compile time.

func (p *PostgresStore) Find(token string) (b []byte, exists bool, err error) {
	panic("missing context arg")
}

func (p *PostgresStore) Commit(token string, b []byte, expiry time.Time) error {
	panic("missing context arg")
}

func (p *PostgresStore) Delete(token string) error {
	panic("missing context arg")
}

func (p *PostgresStore) All() (map[string][]byte, error) {
	panic("missing context arg")
}
