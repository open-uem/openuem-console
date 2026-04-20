package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ali-assar/NATS-Leader-Election/leader"
	"github.com/go-co-op/gocron/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	openuem_nats "github.com/open-uem/nats"
	"github.com/open-uem/openuem-console/internal/controllers/sessions"
	"github.com/open-uem/openuem-console/internal/models"
)

type Handler struct {
	Model *models.Model

	SessionManager       *sessions.SessionManager
	JWTKey               string
	CertPath             string
	KeyPath              string
	SFTPKeyPath          string
	CACertPath           string
	DownloadDir          string
	ServerName           string
	AuthPort             string
	ConsolePort          string
	Domain               string
	TaskScheduler        gocron.Scheduler
	NATSServers          string
	NATSTimeout          int
	NATSConnection       *nats.Conn
	NATSConnectJob       gocron.Job
	JetStream            jetstream.JetStream
	JetStreamCancelFunc  context.CancelFunc
	AgentStream          jetstream.Stream
	ServerStream         jetstream.Stream
	OrgName              string
	OrgProvince          string
	OrgLocality          string
	OrgAddress           string
	Country              string
	ReverseProxyAuthPort string
	ReverseProxyServer   string
	LatestServerRelease  openuem_nats.OpenUEMRelease
	Replicas             int
	ServerReleasesFolder string
	Version              string
	ReenableCertAuth     bool
	ReenablePasswdAuth   bool
	AuthLogger           *log.Logger
	OIDCRedirectURI      string
	CommonAppsJob        gocron.Job
	EncryptionMasterKey  string
}

func NewHandler(model *models.Model, natsServers string, s *sessions.SessionManager, ts gocron.Scheduler, jwtKey, certPath, keyPath, sftpKeyPath, caCertPath, server, consolePort, authPort, tmpDownloadDir, domain, orgName, orgProvince, orgLocality, orgAddress, country, reverseProxyAuthPort, reverseProxyServer, serverReleasesFolder, commonFolder, version, encryptionMasterKey string, reEnableCertAuth, reEnablePasswdAuth bool, authLogger *log.Logger) *Handler {

	// Get NATS request timeout seconds
	timeout, err := model.GetNATSTimeout()
	if err != nil {
		timeout = 20
		log.Println("[ERROR]: could not get NATS request timeout from database")
	}

	// Get Replicas number
	replicas := strings.Split(natsServers, ",")

	h := Handler{
		Model:                model,
		SessionManager:       s,
		JWTKey:               jwtKey,
		CertPath:             certPath,
		KeyPath:              keyPath,
		SFTPKeyPath:          sftpKeyPath,
		CACertPath:           caCertPath,
		DownloadDir:          tmpDownloadDir,
		ServerName:           server,
		ConsolePort:          consolePort,
		AuthPort:             authPort,
		Domain:               domain,
		NATSTimeout:          timeout,
		NATSServers:          natsServers,
		TaskScheduler:        ts,
		OrgName:              orgName,
		OrgProvince:          orgProvince,
		OrgLocality:          orgLocality,
		OrgAddress:           orgAddress,
		Country:              country,
		ReverseProxyAuthPort: reverseProxyAuthPort,
		ReverseProxyServer:   reverseProxyServer,
		Replicas:             len(replicas),
		ServerReleasesFolder: serverReleasesFolder,
		Version:              version,
		ReenableCertAuth:     reEnableCertAuth,
		ReenablePasswdAuth:   reEnablePasswdAuth,
		AuthLogger:           authLogger,
		EncryptionMasterKey:  encryptionMasterKey,
	}

	// Try to create the NATS Connection and start a job if it can't be possible to connect
	if err := h.StartNATSConnectJob(); err != nil {
		log.Fatalf("[FATAL]: could not start NATS Connect job")
	}

	return &h
}

func (h *Handler) StartNATSConnectJob() error {
	var err error
	var ctx context.Context

	h.NATSConnection, err = openuem_nats.ConnectWithNATS(h.NATSServers, h.CertPath, h.KeyPath, h.CACertPath, "")
	if err == nil {
		h.JetStream, err = jetstream.New(h.NATSConnection)
		if err == nil {
			ctx, h.JetStreamCancelFunc = context.WithTimeout(context.Background(), 60*time.Minute)

			agentStreamConfig := jetstream.StreamConfig{
				Name:      "AGENTS_STREAM",
				Subjects:  []string{"agent.certificate.>", "agent.enable.>", "agent.disable.>", "agent.report.>", "agent.update.>", "agent.uninstall.>"},
				Retention: jetstream.InterestPolicy,
			}

			if h.Replicas > 1 {
				agentStreamConfig.Replicas = h.Replicas
			}

			h.AgentStream, err = h.JetStream.CreateOrUpdateStream(ctx, agentStreamConfig)
			if err == nil {
				log.Println("[INFO]: agent stream could be instantiated")

				h.ServerStream, err = h.JetStream.Stream(ctx, "SERVERS_STREAM")
				if err == nil {
					log.Println("[INFO]: server stream could be instantiated")

					// Election
					go func() {
						h.StartAppsDBElection(ctx)
					}()

					return nil
				} else {
					serversExists, err := h.Model.ServersExists()
					if err != nil {
						log.Println("[INFO]: could not check if OpenUEM server exists")
					} else {
						if serversExists {
							log.Printf("[ERROR]: Server Stream could not be instantiated, reason: %v", err)
						}
					}
				}

			} else {
				log.Printf("[ERROR]: Agent Stream could not be instantiated, reason: %v", err)
			}
		} else {
			log.Printf("[ERROR]: could not create Jetstream connection, reason: %v", err)
		}
	} else {
		log.Printf("[ERROR]: could not connect to NATS, reason: %v", err)
	}

	h.NATSConnectJob, err = h.TaskScheduler.NewJob(
		gocron.DurationJob(
			time.Duration(time.Duration(2*time.Minute)),
		),
		gocron.NewTask(
			func() {
				if h.NATSConnection == nil {
					h.NATSConnection, err = openuem_nats.ConnectWithNATS(h.NATSServers, h.CertPath, h.KeyPath, h.CACertPath, "")
					if err != nil {
						log.Printf("[ERROR]: could not connect to NATS %v", err)
						return
					}
				}

				if h.JetStream == nil {
					h.JetStream, err = jetstream.New(h.NATSConnection)
					if err != nil {
						log.Printf("[ERROR]: could not instantiate JetStream, reason: %v", err)
						return
					}
				}

				h.JetStream, err = jetstream.New(h.NATSConnection)
				if err != nil {
					log.Println("[ERROR]: JetStream could not be instantiated")
					return
				}

				ctx, h.JetStreamCancelFunc = context.WithTimeout(context.Background(), 60*time.Minute)

				agentStreamConfig := jetstream.StreamConfig{
					Name:      "AGENTS_STREAM",
					Subjects:  []string{"agent.certificate.>", "agent.enable.>", "agent.disable.>", "agent.report.>", "agent.update.>", "agent.uninstall.>"},
					Retention: jetstream.InterestPolicy,
				}

				if h.Replicas > 1 {
					agentStreamConfig.Replicas = h.Replicas
				}

				h.AgentStream, err = h.JetStream.CreateOrUpdateStream(ctx, agentStreamConfig)
				if err != nil {
					log.Printf("[ERROR]: Agent Stream could not be created or updated, reason: %v", err)
					return
				}

				h.ServerStream, err = h.JetStream.Stream(ctx, "SERVERS_STREAM")
				if err != nil {
					serversExists, err := h.Model.ServersExists()
					if err != nil {
						log.Println("[INFO]: could not check if OpenUEM server exists")
					} else {
						if serversExists {
							log.Printf("[ERROR]: Server Stream could not be created or updated, reason: %v", err)
							return
						}
					}

				}

				if err := h.TaskScheduler.RemoveJob(h.NATSConnectJob.ID()); err != nil {
					return
				}

				// Election
				go func() {
					h.StartAppsDBElection(ctx)
				}()
			},
		),
	)
	if err != nil {
		log.Fatalf("[FATAL]: could not start the NATS connect job: %v", err)
		return err
	}
	log.Printf("[INFO]: new NATS connect job has been scheduled every %d minutes", 2)
	return nil
}

func (h *Handler) StartAppsDBElection(ctx context.Context) {
	// Reference: https://github.com/ali-assar/NATS-Leader-Election/blob/main/cmd/demo/main.go
	// Step 1: Create or get KV bucket
	bucketName := "leaders"
	_, err := h.JetStream.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:  bucketName,
		TTL:     10 * time.Second,
		Storage: jetstream.FileStorage,
	})
	if err != nil {
		// Check if bucket already exists (that's OK, we can use it)
		_, getErr := h.JetStream.KeyValue(ctx, bucketName)
		if getErr != nil {
			log.Fatalf("[FATAL]:Failed to create or get KV bucket for Apps Database election: %v", err)
		}
	}
	log.Printf("[INFO]: key-value bucket %s is ready\n", bucketName)

	// Step 2: Create election configuration
	hostname, _ := os.Hostname()
	cfg := leader.ElectionConfig{
		Bucket:             bucketName,
		Group:              "openuem",
		InstanceID:         fmt.Sprintf("%s-%d", hostname, os.Getpid()),
		TTL:                10 * time.Second,
		HeartbeatInterval:  1 * time.Second,
		ValidationInterval: 5 * time.Second,
	}

	// Step 3: Create election instance
	election, err := leader.NewElectionWithConn(h.NATSConnection, cfg)
	if err != nil {
		log.Fatalf("[FATAL]: Failed to create election: %v", err)
	}
	log.Printf("[INFO]: leader election instance created")

	// Step 4: Register OnPromote callback. This is called when THIS instance becomes leader
	election.OnPromote(func(ctx context.Context, token string) {
		log.Println("[INFO]: this console instance has been promoted to leader")

		// Start a job to create sofwate package table from flatpak, brew, and winget databases
		if err := h.StartCommonPackagesDBJob(); err != nil {
			log.Printf("[ERROR]: could not start job to create common packages db, reason: %s", err.Error())
			return
		}
	})

	// Step 5: Register OnDemote callback. This is called when THIS instance loses leadership
	election.OnDemote(func() {
		log.Println("[INFO]: this console instance has been demoted from leader")

		if err := h.TaskScheduler.RemoveJob(h.CommonAppsJob.ID()); err != nil {
			log.Printf("[ERROR]: could not remove the job that updates the software packages table, reason: %v", err)
			return
		}
	})

	// Step 6: Start the election
	// electionContext := context.Background()
	if err := election.Start(ctx); err != nil {
		log.Fatalf("[FATAL]: Failed to start election for Apps Database: %v", err)
	}
	log.Printf("[INFO]: leader election started")

	// // Step 7: DEBUG: Print status periodically
	// statusTicker := time.NewTicker(3 * time.Second)
	// defer statusTicker.Stop()

	// go func() {
	// 	for {
	// 		select {
	// 		case <-statusTicker.C:
	// 			status := election.Status()
	// 			if status.IsLeader {
	// 				log.Printf("[INFO]: Status: LEADER | ID: %s | Token: %s | State: %s\n",
	// 					status.LeaderID, status.Token, status.State)
	// 			} else {
	// 				log.Printf("[INFO]: Status: FOLLOWER | Current Leader: %s | State: %s\n",
	// 					status.LeaderID, status.State)
	// 			}
	// 		case <-electionContext.Done():
	// 			return
	// 		}
	// 	}
	// }()

	// Step 8: Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	log.Printf("[INFO]: Apps DB elector process received a shutdown signal, stopping gracefully...")

	// Step 9: Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = election.StopWithContext(shutdownCtx, leader.StopOptions{
		DeleteKey:     true, // Delete key for fast failover
		WaitForDemote: true, // Wait for OnDemote callback to complete
	})
	if err != nil {
		log.Printf("[ERROR]: Apps DB elector process shutdown err: %v", err)
	} else {
		log.Printf("[INFO]: Apps DB elector process shutdown complete")
	}
}
