package sessions

import (
	"context"
	"log"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionManager struct {
	Manager *scs.SessionManager
	Pool    *pgxpool.Pool
}

func New(dbUrl string, sessionLifetimeInMinutes int) *SessionManager {
	var err error
	sm := SessionManager{}

	sm.Pool, err = pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Println("[FATAL]: session manager could not contact with the database")
		log.Fatal(err)
	}

	sm.Manager = scs.New()
	sm.Manager.Lifetime = time.Duration(sessionLifetimeInMinutes) * time.Minute
	sm.Manager.Store = pgxstore.New(sm.Pool)
	sm.Manager.Cookie.Secure = true
	return &sm
}

func (s *SessionManager) Close() {
	s.Pool.Close()
}
