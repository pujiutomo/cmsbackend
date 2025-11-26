package database

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/joho/godotenv"
)

type FTPConfig struct {
	FTPHost     string
	FTPUser     string
	FTPPassword string
	FTPPort     string
	FTPDir      string
}

func LoadFTPConfig(host string) (*FTPConfig, error) {
	hst := strings.Split(host, ",")
	ClientEnv := "configclient/.env-" + hst[0]

	if err := godotenv.Load(ClientEnv); err != nil {
		return nil, fmt.Errorf("error loading env file %s: %w", ClientEnv, err)
	}

	config := &FTPConfig{
		FTPHost:     os.Getenv("FTP_HOST"),
		FTPUser:     os.Getenv("FTP_USER"),
		FTPPassword: os.Getenv("FTP_PASS"),
		FTPDir:      os.Getenv("FTP_DIR"),
	}

	if config.FTPPort == "" {
		config.FTPPort = "21"
	}

	// Validate required fields
	if config.FTPHost == "" {
		return nil, fmt.Errorf("FTP_HOST not set in environment file")
	}
	if config.FTPUser == "" {
		return nil, fmt.Errorf("FTP_USER not set in environment file")
	}
	if config.FTPPassword == "" {
		return nil, fmt.Errorf("FTP_PASS not set in environment file")
	}

	return config, nil
}

func ConnectFTP(host string) (*ftp.ServerConn, error) {
	config, err := LoadFTPConfig(host)
	if err != nil {
		return nil, err
	}

	ftpClient, err := ftp.Dial(config.FTPHost+":"+config.FTPPort, ftp.DialWithTimeout(30*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to FTP server %s:%s: %w", config.FTPHost, config.FTPPort, err)
	}

	err = ftpClient.Login(config.FTPUser, config.FTPPassword)
	if err != nil {
		ftpClient.Quit()
		return nil, fmt.Errorf("failed to login to FTP server: %w", err)
	}

	return ftpClient, nil
}
