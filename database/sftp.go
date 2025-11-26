package database

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SFTPConfig struct {
	SftpHost     string
	SftpUser     string
	SftpPassword string
	SftpPort     string
	SftpDir      string
}

func LoadSFTPConfig(host string) (*SFTPConfig, error) {
	hst := strings.Split(host, ".")
	clientEnv := "configclient/.env-" + hst[0]

	if err := godotenv.Load(clientEnv); err != nil {
		return nil, fmt.Errorf("error loading env file %s: %w", clientEnv, err)
	}

	config := &SFTPConfig{
		SftpHost:     os.Getenv("SFTP_HOST"),
		SftpUser:     os.Getenv("SFTP_HOST"),
		SftpPassword: os.Getenv("SFTP_PASS"),
		SftpPort:     os.Getenv("SFTP_PORT"),
		SftpDir:      os.Getenv("SFTP_DIR"),
	}

	// Set default port
	if config.SftpPort == "" {
		config.SftpPort = "22"
	}

	// Validate required fields
	if config.SftpHost == "" {
		return nil, fmt.Errorf("SFTP_HOST not set in environment file")
	}
	if config.SftpUser == "" {
		return nil, fmt.Errorf("SFTP_USER not set in environment file")
	}
	if config.SftpPassword == "" {
		return nil, fmt.Errorf("SFTP_PASS not set in environment file")
	}

	return config, nil
}

func ConnectSftp(host string) (*sftp.Client, error) {
	config, err := LoadSFTPConfig(host)
	if err != nil {
		return nil, err
	}

	// SSH client config
	sshConfig := &ssh.ClientConfig{
		User: config.SftpUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.SftpPassword),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Connect to SSH server
	sshClient, err := ssh.Dial("tcp", config.SftpHost+":"+config.SftpPort, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SSH server %s:%s: %w",
			config.SftpHost, config.SftpPort, err)
	}

	// Create SFTP client
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	return sftpClient, nil
}

func GetSFTPDirectory(host string) (string, error) {
	config, err := LoadSFTPConfig(host)
	if err != nil {
		return "", err
	}
	return config.SftpDir, nil
}
