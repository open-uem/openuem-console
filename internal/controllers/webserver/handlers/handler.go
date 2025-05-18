package handlers

import (
	"context"
	"log"
	"strings"
	"time"

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
	WingetFolder         string
	FlatpakFolder        string
	CommonFolder         string
	Version              string
}

func NewHandler(model *models.Model, natsServers string, s *sessions.SessionManager, ts gocron.Scheduler, jwtKey, certPath, keyPath, sftpKeyPath, caCertPath, server, authPort, tmpDownloadDir, domain, orgName, orgProvince, orgLocality, orgAddress, country, reverseProxyAuthPort, reverseProxyServer, serverReleasesFolder, wingetFolder, flatpakFolder, commonFolder, version string) *Handler {

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
		WingetFolder:         wingetFolder,
		FlatpakFolder:        flatpakFolder,
		CommonFolder:         commonFolder,
		Version:              version,
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

	h.NATSConnection, err = openuem_nats.ConnectWithNATS(h.NATSServers, h.CertPath, h.KeyPath, h.CACertPath)
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
					return nil
				} else {
					log.Printf("[ERROR]: Server Stream could not be instantiated, reason: %v", err)
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
					h.NATSConnection, err = openuem_nats.ConnectWithNATS(h.NATSServers, h.CertPath, h.KeyPath, h.CACertPath)
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
					log.Printf("[ERROR]: Server Stream could not be created or updated, reason: %v", err)
					return
				}

				if err := h.TaskScheduler.RemoveJob(h.NATSConnectJob.ID()); err != nil {
					return
				}
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
